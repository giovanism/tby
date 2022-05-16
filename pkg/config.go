package pkg

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var (
	ErrInvalidConfig     = errors.New("Error invalid config")
	ErrInvalidTunnelType = errors.New("Error invalid tunnel type")
)

type Config struct {
	Tunnels []Tunnel `yaml:"tunnels"`
}

func (t *Config) UnmarshalYAML(val *yaml.Node) error {
	type tmpTbyConfig struct {
		Tunnels []yaml.Node `yaml:"tunnels"`
	}
	var tmpConfig tmpTbyConfig

	err := val.Decode(&tmpConfig)
	if err != nil {
		return err
	}

	tunnels := make([]Tunnel, 0, len(tmpConfig.Tunnels))
	for _, tunNode := range tmpConfig.Tunnels {
		if tunNode.Kind != yaml.MappingNode {
			return ErrInvalidConfig
		}
		var tmpTun TunnelType

		err = tunNode.Decode(&tmpTun)
		if err != nil {
			return err
		}

		switch tmpTun.Type {
		case "ssh":
			var sshTun SSHTunnel

			err = tunNode.Decode(&sshTun)
			if err != nil {
				return err
			}

			tunnels = append(tunnels, sshTun)
		case "k8s":
			var k8sTun K8sPortForwardTunnel

			err = tunNode.Decode(&k8sTun)
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
