package main

import (
	"log"
	"time"

	"github.com/plasma-containers/plasma/cli"
	"github.com/plasma-containers/plasma/controller"
	"github.com/plasma-containers/plasma/db"
	"github.com/plasma-containers/plasma/server"
)

func main() {
	if cli.IsCLI() {
		cli.Run()
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		go server.Run()
		for {
			if db.DB != nil {
				break
			}
			time.Sleep(1 * time.Second)
		}
		go controller.Run()
		c := make(chan int)
		<-c
	}
}
