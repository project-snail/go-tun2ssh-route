package tun_route

import (
	"errors"
	"fmt"
	"github.com/eycorsican/go-tun2socks/common/log"
	yijunjunRoute "github.com/yijunjun/route-table"
	"net"
	"syscall"
	"unsafe"
)

var table, _ = yijunjunRoute.NewRouteTable()

func (route *RouteRow) Add() error {

	defer table.Close()

	intf := MustResolveInterface(net.ParseIP(route.Gateway))

	i, err := GetInterfaceByIndex(uint32(intf.Index))
	if err != nil {
		log.Fatalf("GetInterfaceByIndex failed %v", err)
	}

	row := yijunjunRoute.RouteRow{
		ForwardDest:    yijunjunRoute.Inet_aton(route.Addr, false),
		ForwardMask:    yijunjunRoute.Inet_aton(route.Mask, false),
		ForwardNextHop: yijunjunRoute.Inet_aton(route.Gateway, false),
		ForwardIfIndex: i.InterfaceIndex,
		ForwardType:    3,
		ForwardProto:   3,
		ForwardMetric1: i.Metric,
	}

	if err := table.AddRoute(row); err != nil {
		log.Fatalf(err.Error())
	}

	return nil
}

type IPInterfaceEntry struct {
	Family                               uint32
	InterfaceLuid                        uint64
	InterfaceIndex                       uint32
	MaxReassemblySize                    uint32
	InterfaceIdentifier                  uint64
	MinRouterAdvertisementInterval       uint32
	MaxRouterAdvertisementInterval       uint32
	AdvertisingEnabled                   bool
	ForwardingEnabled                    bool
	WeakHostSend                         bool
	WeakHostReceive                      bool
	UseAutomaticMetric                   bool
	UseNeighborUnreachabilityDetection   bool
	ManagedAddressConfigurationSupported bool
	OtherStatefulConfigurationSupported  bool
	AdvertiseDefaultRoute                bool
	RouterDiscoveryBehavior              uint32
	DadTransmits                         uint32
	BaseReachableTime                    uint32
	RetransmitTime                       uint32
	PathMtuDiscoveryTimeout              uint32
	LinkLocalAddressBehavior             uint32
	LinkLocalAddressTimeout              uint32
	ZoneIndices                          [16]uint32
	SitePrefixLength                     uint32
	Metric                               uint32
	NlMtu                                uint32
	Connected                            bool
	SupportsWakeUpPatterns               bool
	SupportsNeighborDiscovery            bool
	SupportsRouterDiscovery              bool
	ReachableTime                        uint32
	TransmitOffload                      InterfaceOffloadRod
	ReceiveOffload                       InterfaceOffloadRod
	DisableDefaultRoutes                 bool
}

type InterfaceOffloadRod struct {
	ChecksumSupported          bool
	OptionsSupported           bool
	DatagramChecksumSupported  bool
	StreamChecksumSupported    bool
	StreamOptionsSupported     bool
	StreamFastPathCompatible   bool
	DatagramFastPathCompatible bool
	LargeSendOffloadSupported  bool
	GiantSendOffloadSupported  bool
}

func GetInterfaceByIndex(index uint32) (*IPInterfaceEntry, error) {
	ie := &IPInterfaceEntry{
		Family:         2, // AF_INET (IPv4)
		InterfaceIndex: index,
	}

	// https://msdn.microsoft.com/en-us/library/windows/desktop/aa814417(v=vs.85).aspx
	iphlpapi := syscall.MustLoadDLL("iphlpapi.dll")
	defer iphlpapi.Release()
	getIpInterfaceEntry := iphlpapi.MustFindProc("GetIpInterfaceEntry")
	r1, r2, err := getIpInterfaceEntry.Call(uintptr(unsafe.Pointer(ie)))
	log.Debugf("%+v", ie)

	return ie, buildError(r1, r2, err)
}

var systemError = map[uint32]string{
	0:    "ERROR_SUCCESS",
	2:    "ERROR_FILE_NOT_FOUND",
	5:    "ERROR_ACCESS_DENIED",
	50:   "ERROR_NOT_SUPPORTED",
	87:   "ERROR_INVALID_PARAMETER",
	122:  "ERROR_INSUFFICIENT_BUFFER",
	1168: "ERROR_NOT_FOUND",
}

func buildError(r1 uintptr, r2 uintptr, err error) error {
	log.Debugf("r1=%v r2=%v err=%+v", r1, r2, err)
	if r1 == 0 {
		return nil
	}
	if m, ok := systemError[uint32(r1)]; ok {
		return errors.New(m)
	}
	return errors.New(fmt.Sprintf("ERROR CODE %d", r1))
}

func MustResolveInterface(gatewayAddress net.IP) net.Interface {
	i, err := ResolveInterface(gatewayAddress)
	if err != nil {
		log.Fatalf("MustResolveInterface failed %v", err)
	}
	return i
}

func ResolveInterface(gatewayAddress net.IP) (net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return net.Interface{}, err
	}

	var viableInterfaces []net.Interface

	for _, intf := range interfaces {
		// Skip down and loopback interfaces
		if (intf.Flags&net.FlagUp == 0) || (intf.Flags&net.FlagLoopback) != 0 {
			continue
		}

		addrs, err := intf.Addrs()
		if err != nil {
			continue
		}

		var viableGatewayAddress net.IP

		for _, addr := range addrs {
			ipAddr, ok := addr.(*net.IPNet)

			// Skip loopback and link-local addresses
			if !ok || !ipAddr.IP.IsGlobalUnicast() {
				continue
			}

			// Skip IPv6 addresses
			if ipAddr.IP.To4() == nil {
				continue
			}

			// Skip addresses not matching target gateway, if specified
			if gatewayAddress != nil && !ipAddr.IP.Equal(gatewayAddress) {
				continue
			}

			viableGatewayAddress = ipAddr.IP
		}

		// Skip interfaces without a viable gateway address
		if viableGatewayAddress == nil {
			continue
		}

		viableInterfaces = append(viableInterfaces, intf)
	}

	switch len(viableInterfaces) {
	case 1:
		return viableInterfaces[0], nil
	case 0:
		return net.Interface{}, errors.New("No viable interface detected!")
	default:
		return net.Interface{}, errors.New("Multiple viable interfaces detected! Please specify a gateway address.")
	}
}
