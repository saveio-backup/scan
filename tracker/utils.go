package tracker

import "fmt"

var ipAddr [4]byte

func ipconvert(ip []byte) [4]byte {
	fmt.Println(ip)
	// for i := range ip {
	// 	ipAddr[i] = ip[i]
	// }
	// return ipAddr

	copy(ipAddr[:4], ip[:4])
	return ipAddr
}
