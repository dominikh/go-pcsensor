package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"honnef.co/go/pcsensor"

	"github.com/kylelemons/gousb/usb"
)

func main() {
	ctx := usb.NewContext()
	sensors, err := pcsensor.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sig
		for _, sensor := range sensors {
			_ = sensor.Close()
		}
		os.Exit(1)
	}()

	for _, sensor := range sensors {
		for {
			log.Println(sensor.Temperatures())
			time.Sleep(1 * time.Second)
		}
	}
}
