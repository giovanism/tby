package pkg

type Tunnel interface {
	Name() string
	Status() string
	PortMapping() string
	Up() error
	Down() error
	IsUp() bool
	GetLocalPort() int
}

type tunnelType struct {
	Type string `yaml:"type"`
}

