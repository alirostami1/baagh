package main

import (
	"log"
	"time"

	"github.com/AliRostami1/baagh/internal/application"
	"github.com/AliRostami1/baagh/pkg/controller/gpio"
	"github.com/AliRostami1/baagh/pkg/sensor"
)

func main() {
	app, err := application.New()
	if err != nil {
		log.Fatalf("there was a problem initiating the application: %v", err)
	}

	// initialize rpio package and allocate memory
	gpioController, err := gpio.New(app.Ctx, app.Db)
	if err != nil {
		app.Log.Fatalf("there was a problem initiating the gpio controller: %v", err)
	}

	if _, err := gpioController.OutputAlarm(10, "9", 7*time.Second); err != nil {
		app.Log.Fatalf("there was a problem with the gpio controller: %v", err)
	}

	errch := gpioController.Input(9, sensor.PullDown)
	go func() {
		err := <-errch
		app.Log.Fatalf("there was a problem with the gpio controller: %v", err)
	}()

	<-app.Ctx.Done()
}
