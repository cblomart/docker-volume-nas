package main

import (
	"flag"
	"fmt"
	"log"
	"os/user"
	"runtime"
	"strconv"

	"github.com/cblomart/docker-volume-nas/plugin"
	"github.com/docker/go-plugins-helpers/sdk"
	"github.com/docker/go-plugins-helpers/volume"
)

var (
	listento   string
	listenport int
	sysmount   string
)

func init() {
	flag.IntVar(&listenport, "port", 8080, "port to listen to if listening to TCP")
	flag.StringVar(&listento, "type", "socket", "type of listen on 'socket' or 'TCP'")
	flag.StringVar(&sysmount, "sysmp", "/mnt", "system mount point to use as base")
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
	plugin := plugin.Nas{}
	h := volume.NewHandler(&plugin)
	if listento == "TCP" {
		h.ServeTCP(plugin.Name(), fmt.Sprintf("locahost:%d", listenport), sdk.WindowsDefaultDaemonRootDir(), nil)
	} else if listento == "socket" {
		if runtime.GOOS == "linux" {
			u, _ := user.Lookup("root")
			gid, _ := strconv.Atoi(u.Gid)
			h.ServeUnix(plugin.Name(), gid)
		} else if runtime.GOOS == "windows" {
			h.ServeWindows(fmt.Sprintf("\\\\.\\pipe\\%s", plugin.Name()), plugin.Name(), sdk.WindowsDefaultDaemonRootDir(), nil)
		}
	}

}
