package plugin

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"syscall"
)

// Regex for volume names
const nameRegex = "^[A-Za-z0-9-_][A-Za-z0-9-_.]+$"

// GetID returns a gid or uid from string
func GetID(id string) int {
	if runtime.GOOS == "windows" {
		log.Println("setting GID or UID on windows is not supported. defaulting to 0")
		return 0
	}
	oid, err := strconv.Atoi(id)
	if err == nil {
		log.Printf("uid or gid option must be an integer. Defaulting to 0")
		return 0
	}
	return oid
}

// GetGUID return uid and gid from optionts
func GetGUID(options map[string]string) (int, int) {
	if runtime.GOOS == "windows" {
		log.Println("setting GID or UID on windows is not supported. defaulting to 0")
		return 0, 0
	}
	ouid := 0
	ogid := 0
	if suid, ok := options["uid"]; ok {
		ouid = GetID(suid)
	}
	if sgid, ok := options["gid"]; ok {
		ogid = GetID(sgid)
	}
	return ouid, ogid

}

func (p *Nas) verbose(message string) {
	if p.Verbose {
		log.Println(message)
	}
}

func (p *Nas) dump(v interface{}) {
	if p.Verbose {
		buf, err := json.Marshal(v)
		if err == nil {
			log.Println(string(buf))
		}
	}
}

// CheckName Validates the name of a volume for the nas plugin
func CheckName(name string) bool {
	validname := regexp.MustCompile(nameRegex)
	return validname.MatchString(name)
}

// CheckTrackFile checks the existence of a track file and returns its path
func checkTrackFile(path string) (string, error) {
	trackPath := fmt.Sprintf("%s/%s", path, TrackFile)
	// check existance of track file
	_, err := os.Stat(trackPath)
	if err != nil {
		if os.IsNotExist(err) {
			trackFile, err := os.OpenFile(trackPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
			if err != nil {
				return "", err
			}
			trackFile.WriteString("#unkonwn volume\n")
			err = trackFile.Close()
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	return trackPath, nil
}

func createPath(path string, uid, gid int) error {
	err := os.Mkdir(path, 700)
	if err != nil {
		log.Printf("Could not create path %s: %s", path, err)
		return err
	}
	// create track file
	trackPath := fmt.Sprintf("%s/%s", path, TrackFile)
	trackFile, err := os.OpenFile(trackPath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("Could not create track file in path %s: %s", path, err)
		return err
	}
	err = trackFile.Close()
	if err != nil {
		log.Printf("Could not close track file %s: %s", trackPath, err)
		return err
	}
	// set uid and gid
	if (uid != 0 || gid != 0) && runtime.GOOS != "windows" {
		err := syscall.Chown(path, uid, gid)
		if err != nil {
			log.Printf("Could not change owner to %d:%d for %s\n", uid, gid, path)
			return err
		}
	}
	return nil
}

func logClose(v string, f *os.File) {
	err := f.Close()
	if err != nil {
		log.Printf("Cannot close track file for volume %s: %s", v, err)
	}
}
