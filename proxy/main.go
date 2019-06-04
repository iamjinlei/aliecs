package main

import (
	"flag"
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/iamjinlei/ecs"
)

func main() {
	port := flag.String("port", "8080", "proxy listening port")
	flag.Parse()

	ecs.Info("starting proxy on port " + *port)

	proxy := goproxy.NewProxyHttpServer()
	if err := http.ListenAndServe(":"+*port, proxy); err != nil {
		ecs.Error("error running proxy %v", err)
	} else {
		ecs.Info("proxy stopped")
	}
}
