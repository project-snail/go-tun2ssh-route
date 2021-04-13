package tun

import (
	"github.com/eycorsican/go-tun2socks/core"
	"github.com/eycorsican/go-tun2socks/tun"
	"io"
	"log"
)

func OpenTunDevice(name, addr, gw, mask string, dnsServers []string) (io.ReadWriteCloser, error) {

	//copy from go-tun2socks/cmd/tun2socks/main.go

	tunDevice, err := tun.OpenTunDevice(
		name,
		addr, gw, mask,
		dnsServers,
		false,
	)
	if err != nil {
		panic(err)
	}

	// Setup TCP/IP stack.
	lwipWriter := core.NewLWIPStack().(io.Writer)

	// Register an output callback to write packets output from lwip stack to tun
	// device, output function should be set before input any packets.
	core.RegisterOutputFn(func(data []byte) (int, error) {
		return tunDevice.Write(data)
	})

	// Copy packets from tun device to lwip stack, it's the main loop.
	go func() {
		_, err := io.CopyBuffer(lwipWriter, tunDevice, make([]byte, 1500))
		if err != nil {
			log.Fatalf("copying data failed: %v", err)
		}
	}()

	return tunDevice, err

}
