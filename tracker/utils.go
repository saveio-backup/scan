/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-08
 */
package tracker

var ipAddr [4]byte

func ipconvert(ip []byte) [4]byte {
	for i := range ip {
		ipAddr[i] = ip[i]
	}
	return ipAddr
}
