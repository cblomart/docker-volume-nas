package main

import (
	"flag"
	"fmt"
	"log"
	"os/user"
	"runtime"
	"strconv"

	"github.com/cblomart/docker-volume-nas/plugin"
	"github.com/cblomart/go-plugins-helpers/sdk"
	"github.com/cblomart/go-plugins-helpers/volume"
)

var (
	listento   string
	listenport int
	sysmount   string
	verbose    bool
)

func init() {
	flag.IntVar(&listenport, "port", 8080, "port to listen to if listening to TCP")
	flag.StringVar(&listento, "type", "socket", "type of listen on 'socket' or 'TCP'")
	flag.StringVar(&sysmount, "sysmp", "/mnt", "system mount point to use as base")
	flag.BoolVar(&verbose, "verbose", false, "Print verbose output")
	flag.Parse()
}

func main() {
	// pre checks
	if len(sysmount) == 0 {
		log.Fatalln("A system mount point must be indicated")
	}
	if listento == "TCP" && listenport < 1000 {
		log.Fatalf("Listen port %d cannot be less than 1000 (system ports)\n", listenport)
	}
	plugin := plugin.Nas{MountPoint: sysmount, Verbose: verbose}
	h := volume.NewHandler(&plugin)
	if listento == "TCP" {
		address := fmt.Sprintf("localhost:%d", listenport)
		log.Printf("Docker Volume Nas plusing listens to %s\n", address)
		err := h.ServeTCP(plugin.Name(), address, sdk.WindowsDefaultDaemonRootDir(), nil)
		if err != nil {
			log.Fatalf("error: %s\n", err)
		}
	} else if listento == "socket" {
		if runtime.GOOS == "linux" {
			u, _ := user.Lookup("root")
			gid, _ := strconv.Atoi(u.Gid)
			address := fmt.Sprintf("/run/docker/plugins/%s.sock", plugin.Name())
			log.Printf("Docker Volume Nas plusing listens to socket %s\n", address)
			err := h.ServeUnix(address, gid)
			if err != nil {
				log.Fatalf("error: %s\n", err)
			}
		} else if runtime.GOOS == "windows" {
			address := fmt.Sprintf("\\\\.\\pipe\\%s", plugin.Name())
			log.Printf("Docker Volume Nas plusing listens to socket %s\n", address)
			err := h.ServeWindows(address, plugin.Name(), sdk.WindowsDefaultDaemonRootDir(), nil)
			if err != nil {
				log.Fatalf("error: %s\n", err)
			}
		}
	}

}
