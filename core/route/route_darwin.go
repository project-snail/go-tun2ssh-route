package tun_route

import (
	"errors"
	"fmt"
	"github.com/eycorsican/go-tun2socks/common/log"
	"net"
	"os"
	"os/exec"
	"strings"
)

func (route *RouteRow) Add() error {

	if os.Geteuid() != 0 {
		log.Fatalf("must be root to alter routing table")
	}

	ip := net.ParseIP(route.Addr)
	if ip == nil {
		return errors.New("invalid IP address")
	}
	params := fmt.Sprintf("add -net %s -netmask %s %s", route.Addr, route.Mask, route.Gateway)

	out, err := exec.Command("route", strings.Split(params, " ")...).Output()
	if err != nil {
		if len(out) != 0 {
			return errors.New(fmt.Sprintf("%v, output: %s", err, out))
		}
		return err
	}
	log.Infof(string(out))

	return nil
}

