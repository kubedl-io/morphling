package main

import (
	"flag"
	"github.com/alibaba/morphling/console/backend/pkg/client"
	"github.com/alibaba/morphling/console/backend/pkg/routers"
)

var (
	port, host, buildDir *string
)

func init() {
	port = flag.String("port", "8081", "the port to listen to for incoming HTTP connections")
	host = flag.String("host", "0.0.0.0", "the host to listen to for incoming HTTP connections")
	buildDir = flag.String("build-dir", "dist", "the dir of frontend")
}

func main() {
	flag.Parse()

	cmgr := client.Init()

	// r:
	r := routers.InitRouter(cmgr)
	client.Start()
	// r:
	_ = r.Run(":9091")
}
