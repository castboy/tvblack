package utils

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

const (
	xForwardedForName = "X-Forwarded-For"
)

type IP struct {
	localNetworks []*net.IPNet
	localIPNet    *net.IP
	serviceIPNet  *net.IP
}

func NewIP(serviceAddr string) (*IP, error) {
	var ip *IP

	_, serviceNetwork, err := net.ParseCIDR(serviceAddr)
	if err != nil {
		return ip, err
	}

	ip = &IP{}
	ip.localNetworks = make([]*net.IPNet, 0)
	for _, sNetwork := range []string{
		"10.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"172.17.0.0/12",
		"172.18.0.0/12",
		"172.19.0.0/12",
		"172.20.0.0/12",
		"172.21.0.0/12",
		"172.22.0.0/12",
		"172.23.0.0/12",
		"172.24.0.0/12",
		"172.25.0.0/12",
		"172.26.0.0/12",
		"172.27.0.0/12",
		"172.28.0.0/12",
		"172.29.0.0/12",
		"172.30.0.0/12",
		"172.31.0.0/12",
		"192.168.0.0/16",
	} {
		_, network, _ := net.ParseCIDR(sNetwork)
		ip.localNetworks = append(ip.localNetworks, network)
	}

	if ip.localIPNet, err = ip.getLocalIP(ip.localNetworks); err != nil {
		return ip, err
	}

	ip.serviceIPNet, err = ip.getLocalIP([]*net.IPNet{serviceNetwork})
	return ip, err
}

func (ipTools *IP) getLocalIP(networks []*net.IPNet) (*net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		ip, _, _ := net.ParseCIDR(addr.String())

		if ip == nil || ip.To4() == nil {
			continue
		}

		for _, network := range networks {
			if network.Contains(ip) {
				return &ip, nil
			}
		}

	}

	return nil, errors.New("are you connected to the network?")
}

func (ipTools *IP) GetLocalIP() string {
	return ipTools.localIPNet.String()
}

func (ipTools *IP) GetServiceIP() string {
	return ipTools.serviceIPNet.String()
}

func (ipTools *IP) GetClientIP(req *http.Request) string {
	xForwardedFor := req.Header.Get(xForwardedForName)

	if xForwardedFor != "" {
		for _, ip := range strings.Split(xForwardedFor, ",") {
			ip = strings.TrimSpace(ip)
			if ip != "" && !ipTools.IsLocalAddr(ip) {
				return ip
			}
		}
	}

	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	return host
}

func (ipTools *IP) IsLocalAddr(addr string) bool {
	ip := net.ParseIP(addr)
	for _, network := range ipTools.localNetworks {
		if network.Contains(ip) {
			return true
		}
	}

	return ip.IsLoopback()
}
