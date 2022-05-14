package pkg

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
)

type SSHTunnel struct {
	TunnelType `yaml:",inline"`
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

func (t SSHTunnel) Name() string {

	return fmt.Sprintf("%s@%s", t.User, t.NodeName)
}

func (t SSHTunnel) PortMapping() string {

	return fmt.Sprintf("%d:%d", t.LocalPort, t.RemotePort)
}

func (t SSHTunnel) Status() string {

	if t.IsUp() {
		return "up"
	}

	return ""
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

func (t SSHTunnel) GetLocalPort() int {

	return t.LocalPort
}
