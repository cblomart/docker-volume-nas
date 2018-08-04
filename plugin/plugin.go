package plugin

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"

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

// GetMountPoint returns sanitized mount point
func (p *Nas) GetMountPoint() string {
	path := strings.Replace(p.MountPoint, "//", "/", -1)
	path = strings.TrimRight(path, "/")
	return path
}

// CheckVolumePath checks the volume path
func (p *Nas) CheckVolumePath(name string) (string, error) {
	path := fmt.Sprintf("%s/%s", p.GetMountPoint(), name)
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("requested volume %s at path %s is not a directory", name, path)
	}
	return path, nil
}

// Create creates a new volume in the mount point
func (p *Nas) Create(request *volume.CreateRequest) error {
	path := fmt.Sprintf("%s/%s", p.GetMountPoint(), request.Name)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// path does not exist
		err := os.Mkdir(path, 700)
		if err != nil {
			log.Printf("Could not create volume %s in path %s: %s", request.Name, path, err)
			return err
		}
		uid, gid := GetGUID(request.Options)
		if (uid != 0 || gid != 0) && runtime.GOOS != "windows" {
			err := syscall.Chown(path, uid, gid)
			if err != nil {
				return err
			}
		}
		return nil
	}
	if err != nil {
		return err
	}
	// path does exist
	if !info.IsDir() {
		return fmt.Errorf("path %s is a file", path)
	}
	// TODO: check directory owner
	return nil
}

// List lists volumes in the mount point
func (p *Nas) List() (*volume.ListResponse, error) {
	infos, err := ioutil.ReadDir(p.GetMountPoint())
	if err != nil {
		return nil, err
	}
	// prepare response
	response := volume.ListResponse{}
	// count folders
	dircount := 0
	for _, info := range infos {
		if info.IsDir() {
			dircount++
		}
	}
	// fill in response
	response.Volumes = make([]*volume.Volume, dircount)
	dircount = 0
	for _, info := range infos {
		v := volume.Volume{Name: info.Name(), Mountpoint: fmt.Sprintf("%s/%s", p.GetMountPoint(), info.Name())}
		response.Volumes[dircount] = &v
		dircount++
	}
	return &response, nil
}

// Get gets a specific volume
func (p *Nas) Get(request *volume.GetRequest) (*volume.GetResponse, error) {
	path, err := p.CheckVolumePath(request.Name)
	if err != nil {
		return nil, err
	}
	response := volume.GetResponse{
		Volume: &volume.Volume{Name: request.Name, Mountpoint: path},
	}
	return &response, nil
}

// Remove removes a volume from the mount point
func (p *Nas) Remove(request *volume.RemoveRequest) error {
	path, err := p.CheckVolumePath(request.Name)
	if err != nil {
		return err
	}
	return os.RemoveAll(path)
}

// Path returns the path with the mount point
func (p *Nas) Path(request *volume.PathRequest) (*volume.PathResponse, error) {
	path, err := p.CheckVolumePath(request.Name)
	if err != nil {
		return nil, err
	}
	response := volume.PathResponse{
		Mountpoint: path,
	}
	return &response, nil
}

// Mount does nothing as the mount point should already be mounted
func (p *Nas) Mount(request *volume.MountRequest) (*volume.MountResponse, error) {
	path, err := p.CheckVolumePath(request.Name)
	if err != nil {
		return nil, err
	}
	response := volume.MountResponse{
		Mountpoint: path,
	}
	return &response, nil
}

// Unmount does nothing as the mount point should already be mounted
func (p *Nas) Unmount(request *volume.UnmountRequest) error {
	_, err := p.CheckVolumePath(request.Name)
	if err != nil {
		return err
	}
	return nil
}

// Capabilities of the module
func (p *Nas) Capabilities() volume.CapabilitiesResponse {
	response := volume.CapabilitiesResponse{
		Capabilities: volume.Capability{
			Scope: "global",
		},
	}
	return response
}
