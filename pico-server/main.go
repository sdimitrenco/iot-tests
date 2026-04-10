package main

// Flash command:
// tinygo flash -target=pico2-w -stack-size=8kb -scheduler=tasks .

import (
	"machine"
	"time"

	"github.com/soypat/cyw43439"
	"github.com/soypat/cyw43439/examples/cywnet"
	"github.com/soypat/lneto/tcp"
	"github.com/soypat/lneto/x/xnet"
)

const (
	ssid     = "easybell DSL-RM4H"
	password = "32600882067192021602"
)

const listenPort = 80

var (
	ledBlue  = machine.GP15
	ledGreen = machine.GP16
)

func main() {
	time.Sleep(2 * time.Second)
	println("Starting HTTP server on Pico 2 W...")

	ledBlue.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ledGreen.Configure(machine.PinConfig{Mode: machine.PinOutput})

	devcfg := cyw43439.DefaultWifiConfig()

	cystack, err := cywnet.NewConfiguredPicoWithStack(ssid, password, devcfg, cywnet.StackConfig{
		Hostname:    "pico-server",
		MaxTCPPorts: 1,
	})
	if err != nil {
		panic("WiFi setup failed: " + err.Error())
	}

	go loopForeverStack(cystack)

	println("Requesting IP via DHCP...")
	var dhcpResults *xnet.DHCPResults
	for {
		dhcpResults, err = cystack.SetupWithDHCP(cywnet.DHCPConfig{})
		if err == nil {
			break
		}
		println("DHCP retry:", err.Error())
		time.Sleep(3 * time.Second)
	}

	println("Server ready at http://" + dhcpResults.AssignedAddr.String())

	stack := cystack.LnetoStack()

	var conn tcp.Conn
	err = conn.Configure(tcp.ConnConfig{
		RxBuf:             make([]byte, 512),
		TxBuf:             make([]byte, 2048),
		TxPacketQueueSize: 3,
	})
	if err != nil {
		panic(err)
	}

	var buf [512]byte

	for {
		err = stack.ListenTCP(&conn, listenPort)
		if err != nil {
			println("listen error:", err.Error())
			time.Sleep(3 * time.Second)
			conn.Abort()
			continue
		}

		println("Waiting for client...")
		for conn.State().IsPreestablished() {
			time.Sleep(5 * time.Millisecond)
		}

		n, _ := conn.Read(buf[:])
		path := parsePath(buf[:n])

		switch path {
		case "/blue/on":
			ledBlue.High()
		case "/blue/off":
			ledBlue.Low()
		case "/green/on":
			ledGreen.High()
		case "/green/off":
			ledGreen.Low()
		}

		writeResponse(&conn, ledBlue.Get(), ledGreen.Get())
		println("Response sent!")
		time.Sleep(100 * time.Millisecond)
		conn.Close()
		time.Sleep(time.Second)
	}
}

// parsePath extracts the URL path from a raw HTTP request line ("GET /path HTTP/1.1").
func parsePath(req []byte) string {
	start := -1
	for i, b := range req {
		if b == ' ' {
			if start == -1 {
				start = i + 1
			} else {
				return string(req[start:i])
			}
		}
	}
	return "/"
}

func writeResponse(conn *tcp.Conn, blue, green bool) {
	blueColor := "#ccc"
	if blue {
		blueColor = "#00f"
	}
	blueLabel := "OFF"
	if blue {
		blueLabel = "ON"
	}
	blueHref := "/blue/on"
	if blue {
		blueHref = "/blue/off"
	}
	blueBtnText := "Turn ON"
	if blue {
		blueBtnText = "Turn OFF"
	}

	greenColor := "#ccc"
	if green {
		greenColor = "#0f0"
	}
	greenLabel := "OFF"
	if green {
		greenLabel = "ON"
	}
	greenHref := "/green/on"
	if green {
		greenHref = "/green/off"
	}
	greenBtnText := "Turn ON"
	if green {
		greenBtnText = "Turn OFF"
	}

	const header = "HTTP/1.1 200 OK\r\nContent-Type: text/html; charset=utf-8\r\nConnection: close\r\n\r\n"
	const cssOpen = `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Pico LEDs</title>` +
		`<style>body{font-family:sans-serif;text-align:center;padding:2em}` +
		`.dot{width:24px;height:24px;border-radius:50%;display:inline-block;vertical-align:middle;margin-right:8px}` +
		`a{padding:8px 18px;background:#333;color:#fff;text-decoration:none;border-radius:5px;margin:6px}` +
		`</style></head><body><h1>Pico 2 W LED Control</h1>`

	body := cssOpen +
		`<p><span class="dot" style="background:` + blueColor + `"></span>` +
		`<b>Blue LED:</b> ` + blueLabel + ` &nbsp;<a href="` + blueHref + `">` + blueBtnText + `</a></p>` +
		`<p><span class="dot" style="background:` + greenColor + `"></span>` +
		`<b>Green LED:</b> ` + greenLabel + ` &nbsp;<a href="` + greenHref + `">` + greenBtnText + `</a></p>` +
		`</body></html>`

	conn.Write([]byte(header))
	conn.Write([]byte(body))
	conn.Flush()
}

func loopForeverStack(stack *cywnet.Stack) {
	for {
		send, recv, _ := stack.RecvAndSend()
		if send == 0 && recv == 0 {
			time.Sleep(5 * time.Millisecond)
		}
	}
}
