package plugin

import (
	"encoding/json"
	"log"
	"runtime"
	"strconv"
)

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
