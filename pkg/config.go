package pkg

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidConfig     = errors.New("Error invalid config")
	ErrInvalidTunnelType = errors.New("Error invalid tunnel type")
)

type Config struct {
	Tunnels []Tunnel `yaml:"tunnels"`
}

func (t *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tmpTbyConfig struct {
		Tunnels []map[string]interface{} `yaml:"tunnels"`
	}
	var tmpConfig tmpTbyConfig

	err := unmarshal(&tmpConfig)
	if err != nil {
		return err
	}

	tunnels := make([]Tunnel, 0, len(tmpConfig.Tunnels))
	for _, tunMap := range tmpConfig.Tunnels {
		tunType, ok := tunMap["type"].(string)
		if !ok {
			return ErrInvalidConfig
		}

		// Re-marshal and unmarshal to get the proper typed struct
		tunBytes, err := yaml.Marshal(tunMap)
		if err != nil {
			return err
		}

		switch tunType {
		case "ssh":
			var sshTun SSHTunnel
			err = yaml.Unmarshal(tunBytes, &sshTun)
			if err != nil {
				return err
			}
			tunnels = append(tunnels, sshTun)
		case "k8s":
			var k8sTun K8sPortForwardTunnel
			err = yaml.Unmarshal(tunBytes, &k8sTun)
			if err != nil {
				return err
			}
			tunnels = append(tunnels, k8sTun)
		default:
			return ErrInvalidTunnelType
		}
	}

	t.Tunnels = tunnels

	return nil
}

func getConfigDir() string {

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

func getConfigPath() string {
	return filepath.Join(getConfigDir(), "tby.yml")
}

func GetConfig() *Config {

	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal().Msgf("Can't load tby config file: %v", err)
	}

	tbyConfig := &Config{}
	err = yaml.Unmarshal([]byte(data), tbyConfig)
	if err != nil {
		log.Fatal().Msgf("Can't parse tby config file: %v", err)
	}

	log.Debug().Msgf("Loaded: %s", configPath)
	return tbyConfig
}
