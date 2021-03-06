package pkg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	. "github.com/giovanism/tby/pkg"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    Config
		wantErr bool
	}{
		{
			name:    "empty config",
			data:    []byte(``),
			want:    Config{},
			wantErr: false,
		},
		{
			name:    "invalid config",
			data:    []byte(`tunnel`),
			want:    Config{},
			wantErr: true,
		},
		{
			name: "valid ssh tunnel",
			data: []byte(`
tunnels:
- type: ssh
  user: root
  nodeName: db
  remotePort: 5432
  localPort: 5432
`),
			want: Config{
				Tunnels: []Tunnel{
					SSHTunnel{
						TunnelType: TunnelType{
							Type: "ssh",
						},
						User:       "root",
						NodeName:   "db",
						RemotePort: 5432,
						LocalPort:  5432,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid k8s tunnel",
			data: []byte(`
tunnels:
- type: k8s
  context: gke-cluster
  resource_namespace: default
  resource_kind: svc
  resource_name: web-server
  remotePort: 8080
  localPort: 80
`),
			want: Config{
				Tunnels: []Tunnel{
					K8sPortForwardTunnel{
						TunnelType: TunnelType{
							Type: "k8s",
						},
						Context: "gke-cluster",
						ResourceNamespace: "default",
						ResourceKind: "svc",
						ResourceName: "web-server",
						RemotePort: 8080,
						LocalPort:  80,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid tunnel type",
			data: []byte(`
tunnels:
- type: invalid
  user: root
  nodeName: db
  remotePort: 5432
  localPort: 5432
`),
			want:    Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config Config
			err := yaml.Unmarshal(tt.data, &config)

			assert.Equal(t, tt.want, config)

			if (err != nil) != tt.wantErr {
				t.Errorf("yaml.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
