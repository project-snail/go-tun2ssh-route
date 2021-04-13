package tun_route

type Route interface {
	Add()
}

type RouteRow struct {
	Addr, Gateway, Mask string
}