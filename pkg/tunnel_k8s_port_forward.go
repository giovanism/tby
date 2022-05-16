package pkg

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type K8sPortForwardTunnel struct {
	TunnelType        `yaml:",inline"`
	Context           string `yaml:"context"`
	ResourceKind      string `yaml:"resource_kind"`
	ResourceNamespace string `yaml:"resource_namespace"`
	ResourceName      string `yaml:"resource_name"`
	RemotePort        int    `yaml:"remotePort"`
	LocalPort         int    `yaml:"localPort"`
}

func (t K8sPortForwardTunnel) getExecArgs() []string {

	return []string{"kubectl", "port-forward", "--context", t.Context, "-n", t.ResourceNamespace, fmt.Sprintf("%s/%s", t.ResourceKind, t.ResourceName), fmt.Sprintf("%d:%d", t.LocalPort, t.RemotePort)}
}

func (t K8sPortForwardTunnel) runPgrep() (*exec.Cmd, error) {

	execStr := strings.Join(t.getExecArgs(), " ")

	log.Debug().Msgf("Running: pgrep -f '%s'", execStr)

	cmd := exec.Command("pgrep", "-f", execStr)

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func (t K8sPortForwardTunnel) runPkill() (*exec.Cmd, error) {

	execStr := strings.Join(t.getExecArgs(), " ")

	log.Debug().Msgf("Running: pkill -f '%s'", execStr)

	cmd := exec.Command("pkill", "-f", execStr)

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func (t K8sPortForwardTunnel) Name() string {

	return fmt.Sprintf("%s/%s", t.ResourceKind, t.ResourceName)
}

func (t K8sPortForwardTunnel) PortMapping() string {

	return fmt.Sprintf("%d:%d", t.LocalPort, t.RemotePort)
}

func (t K8sPortForwardTunnel) Status() string {

	if t.IsUp() {
		return "up"
	}

	return ""
}

func (t K8sPortForwardTunnel) IsUp() bool {

	_, err := t.runPgrep()
	if err != nil {
		log.Err(err).Msg("Error running pgrep or process not found")
		return false
	}

	return true
}

func (t K8sPortForwardTunnel) Up() error {

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

func (t K8sPortForwardTunnel) Down() error {

	_, err := t.runPkill()
	if err != nil {
		return err
	}

	return nil
}

func (t K8sPortForwardTunnel) GetLocalPort() int {

	return t.LocalPort
}
