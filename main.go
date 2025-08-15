package main

import (
	"log"
	"time"

	"github.com/pgulb/plasma/controller"
	"github.com/pgulb/plasma/db"
	"github.com/pgulb/plasma/server"
)

func main() {
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
