package plugin

const (
	// Name is the plugin name
	Name = "volume-nas"
)

// Nas is a simple nas plugin for docker
type Nas struct {
	MountPoint string
}

// Name returns plugin name
func (p *Nas) Name() string {
	return Name
}
