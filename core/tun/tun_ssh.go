package tun

import (
	"acln.ro/zerocopy"
	"github.com/eycorsican/go-tun2socks/common/log"
	"golang.org/x/crypto/ssh"
	"net"
)

type SshTcpHandler struct {
	Client *ssh.Client
}

func (h *SshTcpHandler) Handle(conn net.Conn, target *net.TCPAddr) error {

	go func() {

		remoteConn, err := h.Client.Dial(
			conn.RemoteAddr().Network(),
			target.String(),
		)

		if err != nil {
			log.Fatalf("dial failed %v", err)
			return
		}

		upCh := make(chan struct{})
		go func() {
			_, _ = zerocopy.Transfer(conn, remoteConn)
			upCh <- struct{}{}
		}()
		_, _ = zerocopy.Transfer(remoteConn, conn)
		defer remoteConn.Close()
		defer conn.Close()
		<-upCh
	}()

	log.Infof("new proxy connection to %v", target)

	return nil
}
