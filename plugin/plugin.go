package plugin

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cblomart/go-plugins-helpers/volume"
)

const (
	// Name is the plugin name
	Name = "nas"
	// TrackFile is the name of the file used to track mounts
	TrackFile = ".track"
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
		uid, gid := GetGUID(request.Options)
		err := createPath(path, uid, gid)
		if err != nil {
			log.Printf("error creating volume %s folder %s: %s\n", request.Name, path, err)
			return err
		}
	} else if err != nil {
		log.Printf("Stat error on path %s: %s\n", path, err)
		return err
	} else if !info.IsDir() {
		log.Printf("path %s is a file\n", path)
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
		path := fmt.Sprintf("%s/%s", p.GetMountPoint(), info.Name())
		_, err := checkTrackFile(path)
		if err != nil {
			log.Printf("Error checking track file for volume %s: %s\n", info.Name(), err)
			continue
		}
		v := volume.Volume{
			Name:       info.Name(),
			Mountpoint: path,
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
		log.Printf("%s error getting volume: %s\n", Name, err)
		return nil, err
	}
	_, err = checkTrackFile(path)
	if err != nil {
		log.Printf("Error checking track file for volume %s: %s\n", request.Name, err)
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
	trackPath := fmt.Sprintf("%s/%s", path, TrackFile)
	info, err := os.Stat(trackPath)
	if err != nil {
		return err
	}
	if info.Size() != 0 {
		return fmt.Errorf("cannot remove volume %s with track file not empty", request.Name)
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
	_, err = checkTrackFile(path)
	if err != nil {
		log.Printf("Error checking track file for volume %s: %s\n", request.Name, err)
		return nil, err
	}
	response := volume.PathResponse{
		Mountpoint: path,
	}
	p.verbose("Returning path")
	p.dump(response)
	return &response, nil
}

// Mount tracks the mount call to prevent removal when a volume is used
func (p *Nas) Mount(request *volume.MountRequest) (*volume.MountResponse, error) {
	log.Printf("%s mount volume %s\n", Name, request.Name)
	if !CheckName(request.Name) {
		return nil, fmt.Errorf("Invalid character in %s", request.Name)
	}
	path, err := p.CheckVolumePath(request.Name)
	if err != nil {
		log.Printf("Could not get path for %s: %s\n", request.Name, err)
		return nil, err
	}
	// add the request to the track file
	trackPath, err := checkTrackFile(path)
	if err != nil {
		log.Printf("Error checking track file for volume %s: %s\n", request.Name, err)
		return nil, err
	}
	// open the file
	trackFile, err := os.OpenFile(trackPath, os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		log.Printf("Cannot open trackfile for volume %s: %s\n", request.Name, err)
		return nil, err
	}
	defer logClose(request.Name, trackFile)
	// read the file and check for id presence
	scanner := bufio.NewScanner(trackFile)
	idfound := false
	for scanner.Scan() {
		if request.ID == scanner.Text() {
			idfound = true
			break
		}
	}
	// if not found add it to file
	if !idfound {
		if _, err = trackFile.WriteString(fmt.Sprintf("%s\n", request.ID)); err != nil {
			log.Printf("Cannot append requestor id to track file for volume %s: %s\n", request.Name, err)
			return nil, err
		}
		err := trackFile.Sync()
		if err != nil {
			log.Printf("Cannot sync track file for volume %s: %s\n", request.Name, err)
			return nil, err
		}
	}
	response := volume.MountResponse{
		Mountpoint: path,
	}
	p.verbose("Mount volume:")
	p.dump(response)
	return &response, nil
}

// Unmount tracks the mount call to prevent removal when a volume is used
func (p *Nas) Unmount(request *volume.UnmountRequest) error {
	log.Printf("%s unmount volume %s\n", Name, request.Name)
	if !CheckName(request.Name) {
		return fmt.Errorf("Invalid character in %s", request.Name)
	}
	path, err := p.CheckVolumePath(request.Name)
	if err != nil {
		log.Printf("Could not check path for %s: %s\n", request.Name, err)
		return err
	}
	// add the request to the track file
	trackPath, err := checkTrackFile(path)
	if err != nil {
		log.Printf("Error checking track file for volume %s: %s\n", request.Name, err)
		return err
	}
	// open the file
	trackFile, err := os.OpenFile(trackPath, os.O_RDWR, 0600)
	if err != nil {
		log.Printf("Cannot open trackfile for volume %s: %s\n", request.Name, err)
		return err
	}
	defer logClose(request.Name, trackFile)
	// remove id from file
	lines := []string{}
	idfound := false
	scanner := bufio.NewScanner(trackFile)
	for scanner.Scan() {
		if scanner.Text() != request.ID {
			lines = append(lines, scanner.Text())
		} else {
			idfound = true
		}
	}
	if !idfound {
		log.Printf("Requestor id not found in track file for volume %s\n", request.Name)
		return nil
	}
	_, err = trackFile.Seek(0, 0)
	if err != nil {
		return err
	}
	// write lines to file
	for _, line := range lines {
		_, err := trackFile.WriteString(fmt.Sprintf("%s\n", line))
		if err != nil {
			return err
		}
	}
	trackFile.Truncate(0)
	err = trackFile.Sync()
	if err != nil {
		log.Printf("Cannot sync track file for volume %s: %s\n", request.Name, err)
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
