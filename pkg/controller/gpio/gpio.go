package gpio

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/stianeikeland/go-rpio/v4"

	"github.com/AliRostami1/baagh/pkg/db"
)

type GPIO struct {
	db         *db.Db
	ctx        context.Context
	outputPins []rpio.Pin
}

type EventHandler func(pin int, val bool)
type EventListener struct {
	Key string
	Fn  EventHandler
}

func New(ctx context.Context, db *db.Db) (*GPIO, error) {
	if err := rpio.Open(); err != nil {
		return nil, fmt.Errorf("can't open and memory map GPIO memory range from /dev/mem: %v", err)
	}
	gpio := &GPIO{
		db:  db,
		ctx: ctx,
	}
	go gpio.cleanup()
	return gpio, nil
}

func (g *GPIO) on(pin int, listen *EventListener) error {
	if listen.Key == fmt.Sprint(pin) {
		return fmt.Errorf("circular dependency: pin%[1]o can't depend on pin%[1]o", pin)
	}

	g.db.On(listen.Key, func(key string, val *redis.StringCmd) error {
		v, err := val.Bool()
		if err != nil {
			return fmt.Errorf("can't sync ")
		}
		listen.Fn(pin, v)
		return nil
	})

	return nil
}

func (g *GPIO) Set(pin int, val bool) error {
	p := rpio.Pin(pin)

	if _, err := g.db.Set(fmt.Sprint(pin), val, 0); err != nil {
		return err
	}
	if val {
		p.Write(rpio.High)
	} else {
		p.Write(rpio.Low)
	}
	return nil
}

func (g *GPIO) cleanup() {
	defer rpio.Close()
	<-g.ctx.Done()
	for _, pin := range g.outputPins {
		pin.Low()
	}
}

func (g *GPIO) addOutputPins(pin rpio.Pin) error {
	for _, p := range g.outputPins {
		if p == pin {
			return fmt.Errorf("can't add 2 controllers for the same pin")
		}
	}
	g.outputPins = append(g.outputPins, pin)
	return nil
}
