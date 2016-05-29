package main

import (
	"log"

	"honnef.co/go/pcsensor"

	"github.com/kylelemons/gousb/usb"
)

func main() {
	ctx := usb.NewContext()
	sensors, err := pcsensor.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, sensor := range sensors {
		defer sensor.Close()
		log.Println(sensor.Temperatures())
	}
}
