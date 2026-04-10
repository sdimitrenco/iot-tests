package main

// Flash command:
// tinygo flash -target=pico2-w -stack-size=8kb -scheduler=tasks .

import (
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

func main() {
	time.Sleep(2 * time.Second)
	println("Starting HTTP server on Pico 2 W...")

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

		conn.Read(buf[:])

		const body = "<h1>Hello from Pico 2 W!</h1><p>WiFi is working!</p>"
		const response = "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/html; charset=utf-8\r\n" +
			"Connection: close\r\n" +
			"\r\n" +
			body

		conn.Write([]byte(response))
		conn.Flush()
		println("Response sent!")
		time.Sleep(100 * time.Millisecond)
		conn.Close()
		time.Sleep(time.Second)
	}
}

func loopForeverStack(stack *cywnet.Stack) {
	for {
		send, recv, _ := stack.RecvAndSend()
		if send == 0 && recv == 0 {
			time.Sleep(5 * time.Millisecond)
		}
	}
}
