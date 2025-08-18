package main

import (
	"log"
	"time"

	"github.com/pgulb/plasma/cli"
	"github.com/pgulb/plasma/controller"
	"github.com/pgulb/plasma/db"
	grpcserver "github.com/pgulb/plasma/grpc_server"
	"github.com/pgulb/plasma/server"
	"github.com/pgulb/plasma/version"
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
