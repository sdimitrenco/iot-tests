package main

import (
	"machine"
	"time"
)

func main() {
	led := machine.GP15
    led2 := machine.GP16
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led2.Configure(machine.PinConfig{Mode: machine.PinOutput})

	println("--- Программа запущена! ---")

	for {
		for i := uint32(0); i < 65535; i += 500 {
			led.Set(ch, i)
			time.Sleep(time.Millisecond * 10)
		}
		// Постепенно уменьшаем
		for i := uint32(65535); i > 0; i -= 500 {
			led.Set(ch, i)
			time.Sleep(time.Millisecond * 10)
		}
	}
}
