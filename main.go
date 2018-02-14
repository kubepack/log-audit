package main

import (
	"os"
	"runtime"

	"github.com/golang/glog"
	"github.com/kubepack/packserver/pkg/cmds"
	"k8s.io/apiserver/pkg/util/logs"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	if err := cmds.NewRootCmd().Execute(); err != nil {
		glog.Fatalln("Error in packserver:", err)
	}
	glog.Infoln("Exiting packserver")
	os.Exit(0)
}
