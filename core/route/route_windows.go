package tun_route

import (
	"github.com/eycorsican/go-tun2socks/common/log"
	yijunjunRoute "github.com/yijunjun/route-table"
)

var table, _ = yijunjunRoute.NewRouteTable()

func (route *RouteRow) Add() error {

	defer table.Close()

	rows, err := table.Routes()
	if err != nil {
		panic(err.Error())
	}

	localIp, err := yijunjunRoute.GetLocalIp()
	if err != nil {
		panic(err.Error())
	}

	localIpUint := yijunjunRoute.Inet_aton(localIp, false)

	for _, row := range rows {
		if row.ForwardNextHop == localIpUint {
			// route add route.Addr mask route.Mask route.Gateway
			row.ForwardDest = yijunjunRoute.Inet_aton(route.Addr, false)
			row.ForwardMask = yijunjunRoute.Inet_aton(route.Mask, false)
			row.ForwardNextHop = yijunjunRoute.Inet_aton(route.Gateway, false)
			// 需要管理员权限,才能添加成功
			if err := table.AddRoute(row); err != nil {
				log.Fatalf(err.Error())
			}
			break
		}
	}

	return nil
}
