package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/EugeneKrivoshein/wb_l0/api"
	"github.com/EugeneKrivoshein/wb_l0/cmd/config"
	"github.com/EugeneKrivoshein/wb_l0/internal/db"
	"github.com/EugeneKrivoshein/wb_l0/internal/streaming"
)

func main() {

	config.Config()
	dbObject := db.NewDB()
	csh := db.NewCache(dbObject)
	sh := streaming.NewStreamingHandler(dbObject)
	myApi := api.NewApi(csh)

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			fmt.Printf("\nunsubscribe and closing conn...\n\n")
			csh.Finish()
			sh.Finish()
			myApi.Finish()

			cleanupDone <- true
		}
	}()
	<-cleanupDone
}
