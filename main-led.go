package main

import (
    "machine"
    "time"
)

func main() {
    led := machine.LED
    led.Configure(machine.PinConfig{Mode: machine.PinOutput})

    for {
        println("I am working!") // Это отправится в USB порт
        led.High()
        time.Sleep(time.Second)
        led.Low()
        time.Sleep(time.Second)
    }
}