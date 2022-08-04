package pkg

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type SSHTunnel struct {
	TunnelType `yaml:",inline"`
	User       string `yaml:"user"`
	NodeName   string `yaml:"node_name"`
	RemotePort int    `yaml:"remote_port"`
	LocalPort  int    `yaml:"local_port"`
}

func (t *SSHTunnel) UnmarshalYAML(val *yaml.Node) error {

	type proxyType SSHTunnel
	err := val.Decode((*proxyType)(t))
	if err != nil {
		return err
	}

	type oldSSHTunnelFields struct {
		NodeName   string `yaml:"nodeName"`
		RemotePort int    `yaml:"remotePort"`
		LocalPort  int    `yaml:"localPort"`
	}
	var oldTun oldSSHTunnelFields

	err = val.Decode(&oldTun)
	if err != nil {
		return err
	}

	if t.NodeName == "" && oldTun.NodeName != "" {
		t.NodeName = oldTun.NodeName
	}

	if t.RemotePort == 0 && oldTun.RemotePort != 0 {
		t.RemotePort = oldTun.RemotePort
	}

	if t.LocalPort == 0 && oldTun.LocalPort != 0 {
		t.LocalPort = oldTun.LocalPort
	}

	return nil
}

func (t SSHTunnel) getExecArgs() []string {

	return []string{"tsh", "ssh", "-NL", fmt.Sprintf("%d:localhost:%d", t.LocalPort, t.RemotePort), fmt.Sprintf("%s@%s", t.User, t.NodeName)}
}

func (t SSHTunnel) runPgrep() (*exec.Cmd, error) {

	execStr := strings.Join(t.getExecArgs(), " ")

	log.Debug().Msgf("Running: pgrep -f '%s'", execStr)

	cmd := exec.Command("pgrep", "-f", execStr)

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func (t SSHTunnel) runPkill() (*exec.Cmd, error) {

	execStr := strings.Join(t.getExecArgs(), " ")

	log.Debug().Msgf("Running: pkill -f '%s'", execStr)

	cmd := exec.Command("pkill", "-f", execStr)

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
		log.Warn().AnErr("error", err).Msg("Error running pgrep or process not found")
		return false
	}

	return true
}

func (t SSHTunnel) Up() error {

	execArgs := t.getExecArgs()
	log.Debug().Msgf("Starting: %s", strings.Join(execArgs, " "))

	cmd := exec.Command(execArgs[0], execArgs[1:]...)

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
