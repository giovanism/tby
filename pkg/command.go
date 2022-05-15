package pkg

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func CommandExecute() {
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

			config := GetConfig()
			tun := config.Tunnels[id]

			if tun.IsUp() {
				// Intended idempotency
				log.Warn().Msgf("Tunnel %d on port %d is already up", id, tun.GetLocalPort())
				return
			}

			log.Info().Msgf("Connecting to tunnel %d on port %d", id, tun.GetLocalPort())
			err = tun.Up()
			if err != nil {
				log.Fatal().Err(err).Msgf("Failed to connect to tunnel %d on port %d", id, tun.GetLocalPort())
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

			config := GetConfig()
			tun := config.Tunnels[id]

			log.Info().Msgf("Disconnecting tunnel %d on port %d", id, tun.GetLocalPort())
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

			config := GetConfig()

			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "Id\tName\tPort\tStatus")
			for i, tun := range config.Tunnels {
				fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", i, tun.Name(), tun.PortMapping(), tun.Status())
			}
			tw.Flush()
		},
	}
}
