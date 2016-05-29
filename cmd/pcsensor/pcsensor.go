package main

import (
	"log"

	"github.com/kylelemons/gousb/usb"
	"honnef.co/go/pcsensor"
)

func main() {
	ctx := usb.NewContext()
	sensor, err := pcsensor.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer sensor.Close()
	log.Println(sensor.Temperatures())
}
