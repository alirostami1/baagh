package gpio

import (
	"fmt"
	"sync"
	"time"

	"github.com/AliRostami1/baagh/pkg/debounce"
	"github.com/warthog618/gpiod"
)

type OutputObject struct {
	*Object
}

type OutputOption struct {
	Meta
}

type ObjectEvent struct {
	Key, Value string
	Object     *Object
}

type EventHandler func(item *ObjectEvent)

func (o *OutputObject) On(key string, fns ...EventHandler) error {
	if key == o.key {
		return CircularDependency{key: o.key}
	}

	for _, fn := range fns {
		o.Gpio.db.On(key, func(key, value string) {
			fn(&ObjectEvent{
				Key:    key,
				Value:  value,
				Object: o.Object,
			})

		})
	}

	return nil
}

func (o *OutputObject) OnItem(object *Object, fns ...EventHandler) error {
	return o.On(object.key, fns...)
}

func (o *OutputObject) OnPin(pin int, fns ...EventHandler) error {
	item, err := o.Gpio.getItem(pin)
	if err != nil {
		return err
	}
	return o.OnItem(item, fns...)
}

func (g *Gpio) Output(pin int, opt OutputOption) (*OutputObject, error) {
	outputPin, err := g.chip.RequestLine(pin, gpiod.AsOutput(int(INACTIVE)))
	if err != nil {
		return nil, fmt.Errorf("there was a problem with output controller: %v", err)
	}

	outputInfo, err := outputPin.Info()
	if err != nil {
		return nil, fmt.Errorf("there was a problem with output controller: %v", err)
	}

	output := OutputObject{
		Object: &Object{
			Gpio: g,
			Line: outputPin,
			data: &ObjectData{
				Info:  outputInfo,
				State: 0,
				Meta:  opt.Meta,
			},
			key: makeKey(pin),
			mu:  &sync.RWMutex{},
		},
	}

	g.chip.WatchLineInfo(pin, func(lice gpiod.LineInfoChangeEvent) {
		output.set(func(trx *ObjectTrx) error {
			trx.SetInfo(lice.Info)
			return nil
		})
	})

	err = g.addItem(pin, output.Object)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (g *Gpio) OutputSync(pin int, key string, options OutputOption) (*OutputObject, error) {
	output, err := g.Output(pin, options)
	if err != nil {
		return nil, err
	}
	err = output.On(key, func(evt *ObjectEvent) {
		evt.Object.set(func(trx *ObjectTrx) error {
			trx.SetState(evt.Object.data.State ^ 1)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (g *Gpio) OutputRSync(pin int, key string, options OutputOption) (*OutputObject, error) {
	output, err := g.Output(pin, options)
	if err != nil {
		return nil, err
	}
	err = output.On(key, func(evt *ObjectEvent) {
		evt.Object.set(func(trx *ObjectTrx) error {
			trx.SetState(evt.Object.data.State)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (g *Gpio) OutputAlarm(pin int, key string, delay time.Duration, option OutputOption) (*OutputObject, error) {
	output, err := g.Output(pin, option)
	if err != nil {
		return nil, err
	}
	fn := debounce.Debounce(delay, func() {
		output.set(func(trx *ObjectTrx) error {
			trx.SetState(INACTIVE)
			return nil
		})
	})

	err = output.On(key, func(obj *ObjectEvent) {
		output.set(func(trx *ObjectTrx) error {
			trx.SetState(obj.Object.data.State)
			return nil
		})
		fn()
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}
