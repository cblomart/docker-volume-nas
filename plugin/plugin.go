package plugin

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/cblomart/go-plugins-helpers/volume"
)

const (
	// Name is the plugin name
	Name = "nas"
)

// Nas is a simple nas plugin for docker
type Nas struct {
	MountPoint string
	Verbose    bool
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
	if !CheckName(name) {
		return "", fmt.Errorf("Invalid character in %s", name)
	}
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
	log.Printf("%s create volume %s\n", Name, request.Name)
	if !CheckName(request.Name) {
		return fmt.Errorf("Invalid character in %s", request.Name)
	}
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
				log.Printf("Could not change owner to %d:%d for %s\n", uid, gid, path)
				return err
			}
		}
		return nil
	}
	if err != nil {
		log.Printf("Stat error on path %s: %s", path, err)
		return err
	}
	// path does exist
	if !info.IsDir() {
		log.Printf("path %s is a file", path)
		return fmt.Errorf("path %s is a file", path)
	}
	p.verbose(fmt.Sprintf("Created volume %s with path %s", request.Name, path))
	// TODO: check directory owner
	return nil
}

// List lists volumes in the mount point
func (p *Nas) List() (*volume.ListResponse, error) {
	log.Printf("%s list volumes\n", Name)
	infos, err := ioutil.ReadDir(p.GetMountPoint())
	if err != nil {
		log.Printf("Could not read dir %s\n", err)
		return nil, err
	}
	// prepare response
	response := volume.ListResponse{}
	// fill in response
	response.Volumes = []*volume.Volume{}
	for _, info := range infos {
		if !info.IsDir() {
			p.verbose(fmt.Sprintf("Ignoring file %s", info.Name()))
			continue
		}
		if !CheckName(info.Name()) {
			p.verbose(fmt.Sprintf("Ignoring folder with invalid charachter %s", info.Name()))
			continue
		}
		v := volume.Volume{
			Name:       info.Name(),
			Mountpoint: fmt.Sprintf("%s/%s", p.GetMountPoint(), info.Name()),
		}
		response.Volumes = append(response.Volumes, &v)
	}
	p.verbose("Generated Volume list:")
	p.dump(response)
	return &response, nil
}

// Get gets a specific volume
func (p *Nas) Get(request *volume.GetRequest) (*volume.GetResponse, error) {
	log.Printf("%s get volume %s\n", Name, request.Name)
	if !CheckName(request.Name) {
		return nil, fmt.Errorf("Invalid character in %s", request.Name)
	}
	path, err := p.CheckVolumePath(request.Name)
	if err != nil {
		log.Printf("%s error getting volume: %s", Name, err)
		return nil, err
	}
	response := volume.GetResponse{
		Volume: &volume.Volume{
			Name:       request.Name,
			Mountpoint: path,
		},
	}
	p.verbose("Returning volume:")
	p.dump(response)
	return &response, nil
}

// Remove removes a volume from the mount point
func (p *Nas) Remove(request *volume.RemoveRequest) error {
	log.Printf("%s remove volume %s\n", Name, request.Name)
	if !CheckName(request.Name) {
		return fmt.Errorf("Invalid character in %s", request.Name)
	}
	path, err := p.CheckVolumePath(request.Name)
	if err != nil {
		log.Printf("Could not check path for %s: %s\n", request.Name, err)
		return err
	}
	return os.RemoveAll(path)
}

// Path returns the path with the mount point
func (p *Nas) Path(request *volume.PathRequest) (*volume.PathResponse, error) {
	log.Printf("%s volume path  %s\n", Name, request.Name)
	if !CheckName(request.Name) {
		return nil, fmt.Errorf("Invalid character in %s", request.Name)
	}
	path, err := p.CheckVolumePath(request.Name)
	if err != nil {
		log.Printf("Could not check path for %s: %s\n", request.Name, err)
		return nil, err
	}
	response := volume.PathResponse{
		Mountpoint: path,
	}
	p.verbose("Returning path")
	p.dump(response)
	return &response, nil
}

// Mount does nothing as the mount point should already be mounted
func (p *Nas) Mount(request *volume.MountRequest) (*volume.MountResponse, error) {
	log.Printf("%s mount volume %s\n", Name, request.Name)
	if !CheckName(request.Name) {
		return nil, fmt.Errorf("Invalid character in %s", request.Name)
	}
	path, err := p.CheckVolumePath(request.Name)
	if err != nil {
		log.Printf("Could not get path for %s: %s", request.Name, err)
		return nil, err
	}
	response := volume.MountResponse{
		Mountpoint: path,
	}
	p.verbose("Mount volume:")
	p.dump(response)
	return &response, nil
}

// Unmount does nothing as the mount point should already be mounted
func (p *Nas) Unmount(request *volume.UnmountRequest) error {
	log.Printf("%s unmount volume %s\n", Name, request.Name)
	if !CheckName(request.Name) {
		return fmt.Errorf("Invalid character in %s", request.Name)
	}
	_, err := p.CheckVolumePath(request.Name)
	if err != nil {
		log.Printf("Could not check path for %s: %s\n", request.Name, err)
		return err
	}
	return nil
}

// Capabilities of the module
func (p *Nas) Capabilities() *volume.CapabilitiesResponse {
	log.Printf("%s capabilities\n", Name)
	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{
			Scope: "global",
		},
	}
}
