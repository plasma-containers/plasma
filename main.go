package main

import (
	"log"
	"time"

	"github.com/plasma-containers/plasma/cli"
	"github.com/plasma-containers/plasma/controller"
	"github.com/plasma-containers/plasma/db"
	grpcserver "github.com/plasma-containers/plasma/grpc_server"
	"github.com/plasma-containers/plasma/server"
	"github.com/plasma-containers/plasma/version"
)

func main() {
	if cli.IsCLI() {
		cli.Run()
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Plasma version:", version.Version)
		go server.Run()
		for {
			if db.DB != nil {
				break
			}
			time.Sleep(1 * time.Second)
		}
		go controller.Run()
		go grpcserver.Run()
		c := make(chan int)
		<-c
	}
}
