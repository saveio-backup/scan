package tracker

import "net"

var ipAddr [4]byte

func ipconvert(ip net.IP) [4]byte {
	if len(ip) != 4 {
		ip = ip.To4()
	}
	copy(ipAddr[:4], ip[:4])
	return ipAddr
}
