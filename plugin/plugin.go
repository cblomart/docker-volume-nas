package plugin

import (
	"github.com/docker/go-plugins-helpers/volume"
)

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

// Create creates a new volume in the mount point
func (p *Nas) Create(*volume.CreateRequest) error {

}

// List lists volumes in the mount point
func (p *Nas) List() (*volume.ListResponse, error) {

}

// Get gets a specific volume
func (p *Nas) Get(*volume.GetRequest) (*volume.GetResponse, error) {

}

// Remove removes a volume from the mount point
func (p *Nas) Remove(*volume.RemoveRequest) error {

}

// Path returns the path with the mount point
func (p *Nas) Path(*volume.PathRequest) (*volume.PathResponse, error) {

}

// Mount does nothing as the mount point should already be mounted
func (p *Nas) Mount(*volume.MountRequest) (*volume.MountResponse, error) {

}

// Unmount does nothing as the mount point should already be mounted
func (p *Nas) Unmount(*volume.UnmountRequest) error {

}

// Capabilities of the module
func (p *Nas) Capabilities() *volume.CapabilitiesResponse {

}
