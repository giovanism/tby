package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cheynewallace/tabby"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

var (
	configPath = getTbyConfigPath()
)

type TbyConfig struct {
	Tunnels []SSHTunnel `yaml:"tunnels"`
}

type SSHTunnel struct {
	Type       string `yaml:"type"`
	User       string `yaml:"user"`
	NodeName   string `yaml:"nodeName"`
	RemotePort int    `yaml:"remotePort"`
	LocalPort  int    `yaml:"localPort"`
}

func (t SSHTunnel) runPgrep() (*exec.Cmd, error) {

	log.Debug().Msgf("Running: pgrep -f %s", fmt.Sprintf("%d:localhost:%d", t.LocalPort, t.RemotePort))

	cmd := exec.Command("pgrep", "-f", fmt.Sprintf("%d:localhost:%d", t.LocalPort, t.RemotePort))

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func (t SSHTunnel) runPkill() (*exec.Cmd, error) {

	log.Debug().Msgf("Running: pkill -f %s", fmt.Sprintf("%d:localhost:%d", t.LocalPort, t.RemotePort))

	cmd := exec.Command("pkill", "-f", fmt.Sprintf("%d:localhost:%d", t.LocalPort, t.RemotePort))

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func (t SSHTunnel) IsUp() bool {

	_, err := t.runPgrep()
	if err != nil {
		log.Err(err).Msg("Error running pgrep or process not found")
		return false
	}

	return true
}

func (t SSHTunnel) Up() error {

	log.Debug().Msgf("Starting: tsh ssh -NL %s %s", fmt.Sprintf("%d:localhost:%d", t.LocalPort, t.RemotePort), fmt.Sprintf("%s@%s", t.User, t.NodeName))

	cmd := exec.Command("tsh", "ssh", "-NL", fmt.Sprintf("%d:localhost:%d", t.LocalPort, t.RemotePort), fmt.Sprintf("%s@%s", t.User, t.NodeName))

	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error)

	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		return err
	case <-time.After(1 * time.Second):
		return nil
	}
}

func (t SSHTunnel) Down() error {

	_, err := t.runPkill()
	if err != nil {
		return err
	}

	return nil
}

func getTbyDir() string {

	var configDir, homeDir string

	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, err = os.UserHomeDir()

		if err != nil {
			log.Fatal().Msgf("Can't find user config directory: %v", err)
		}

		return filepath.Join(homeDir, ".tby")
	}

	return filepath.Join(configDir, "tby")
}

func getTbyConfigPath() string {
	return filepath.Join(getTbyDir(), "tby.yml")
}

func getTbyConfig() *TbyConfig {

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal().Msgf("Can't load tby config file: %v", err)
	}

	tbyConfig := &TbyConfig{}
	err = yaml.Unmarshal([]byte(data), tbyConfig)
	if err != nil {
		log.Fatal().Msgf("Can't parse tby config file: %v", err)
	}

	log.Debug().Msgf("Loaded: %s", configPath)
	return tbyConfig
}

func main() {
	rootCmd().Execute()
}

func rootCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "tby ID",
		Short: "Connect to tunnel ID",
		Long: `tby is the main command, used to connect to your tunnels.

tby: teleport behind you
An awesome terminal program that will accelerate your way of using tsh teleport client.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			defer func() {
				if err := recover(); err != nil {
					log.Fatal().Msgf("recovered from panic: %v", err)
				}
			}()

			id, err := strconv.Atoi(args[0])
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to parse ID")
			}

			tbyConfig := getTbyConfig()
			tun := tbyConfig.Tunnels[id]

			log.Info().Msgf("Connecting to tunnel %d on port %d", id, tun.LocalPort)
			err = tun.Up()
			if err != nil {
				log.Fatal().Err(err).Msgf("Failed to connect to tunnel %d on port %d", id, tun.LocalPort)
			}
		},
	}

	cmd.AddCommand(downCmd(), listCmd())

	return cmd
}

func downCmd() *cobra.Command {

	return &cobra.Command{
		Use:   "down ID",
		Short: "Deactivate active tunnel",
		Long:  `Deactivate active tunnel managed by tby.`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			defer func() {
				if err := recover(); err != nil {
					log.Fatal().Msgf("recovered from panic: %v", err)
				}
			}()

			id, err := strconv.Atoi(args[0])
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to parse ID")
			}

			tbyConfig := getTbyConfig()
			tun := tbyConfig.Tunnels[id]

			log.Info().Msgf("Disconnecting tunnel %d on port %d", id, tun.LocalPort)
			err = tun.Down()
			if err != nil {
				log.Fatal().Err(err).Msgf("Failed to deactivate tunnel %d", id)
			}
		},
	}
}

func listCmd() *cobra.Command {

	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List registered tunnels",
		Long:    `List tunnels configured inside tby config file in a table.`,
		Run: func(cmd *cobra.Command, args []string) {

			tbyConfig := getTbyConfig()

			t := tabby.New()
			t.AddHeader("Id", "user@host", "Port", "Status")
			for i, tun := range tbyConfig.Tunnels {
				t.AddLine(i, fmt.Sprintf("%s@%s", tun.User, tun.NodeName), fmt.Sprintf("%d:%d", tun.LocalPort, tun.RemotePort), tun.IsUp())
			}
			t.Print()
		},
	}
}
