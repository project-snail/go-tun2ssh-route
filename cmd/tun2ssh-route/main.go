package main

import (
	"flag"
	"fmt"
	"github.com/eycorsican/go-tun2socks/common/log"
	_ "github.com/eycorsican/go-tun2socks/common/log/simple" // Register a simple logger.
	"github.com/eycorsican/go-tun2socks/core"
	tunRoute "github.com/project-snail/go-tun2ssh-route/core/route"
	"github.com/project-snail/go-tun2ssh-route/core/tun"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	configPath := flag.String("config", "config.yml", "config path")
	flag.Parse()
	content, err := ioutil.ReadFile(*configPath)

	if err != nil {
		log.Fatalf("读取配置文件失败 %v", err)
	}

	tunInfo := TunInfo{}
	err = yaml.Unmarshal(content, &tunInfo)
	if err != nil {
		log.Fatalf("解析配置文件失败 %v", err)
	}

	//注册ssh转发的处理器
	registerSshConnHandler(tunInfo.Ssh)

	ip := net.ParseIP(tunInfo.TunAddr)

	if ip[15] != 0 {
		log.Fatalf("tun设备IP段错误，最后一段必须为0")
	}

	ip[15] = 1
	ip.To4()
	//打开tun设备
	_, err = tun.OpenTunDevice(
		tunInfo.TunName,
		ip.String(),
		tunInfo.TunAddr,
		"255.255.255.0",
		strings.Split(tunInfo.DnsServers, ","),
	)
	if err != nil {
		log.Fatalf("OpenTunDevice失败 %v", err)
	}

	//添加路由
	addRoute(tunInfo.RouteArr, ip.String())

	log.Infof("Running go-tun2ssh-route")

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	<-osSignals

}

func addRoute(routeArr []RouteInfo, gateway string) {

	for _, route := range routeArr {
		r := tunRoute.RouteRow{
			Addr:    route.Addr,
			Gateway: gateway,
			Mask:    route.Mask,
		}
		err := r.Add()
		if err != nil {
			log.Fatalf("add route failed %v", err)
		}
	}

}

func registerSshConnHandler(sshInfo SshInfo) {

	var authMethod ssh.AuthMethod

	switch sshInfo.AuthType {
	case "password":
		authMethod = ssh.Password(sshInfo.Password)
	case "identity":
		key, err := ioutil.ReadFile(sshInfo.Password)
		if err != nil {
			log.Fatalf("Read identity file failed %v", err)
		}
		privateKey, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(sshInfo.Passphrase))
		if err != nil {
			log.Fatalf("Parse private key failed %v", err)
		}
		authMethod = ssh.PublicKeys(privateKey)
	}

	client, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", sshInfo.Host, sshInfo.Port),
		&ssh.ClientConfig{
			User:            sshInfo.UserName,
			Auth:            []ssh.AuthMethod{authMethod},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	)

	if err != nil {
		panic(err)
	}

	core.RegisterTCPConnHandler(&tun.SshTcpHandler{Client: client})
}

type RouteInfo struct {
	Addr string
	Mask string
}

type SshInfo struct {
	Host       string
	Port       int32
	UserName   string `yaml:"user-name"`
	AuthType   string `yaml:"auth-type"`
	Password   string
	Passphrase string
}

type TunInfo struct {
	TunName    string `yaml:"tun-name"`
	TunAddr    string `yaml:"tun-addr"` // tun设备将使用的ip段 10.255.123.0
	DnsServers string `yaml:"dns-servers"`
	Ssh        SshInfo
	RouteArr   []RouteInfo `yaml:"route-arr"`
}
