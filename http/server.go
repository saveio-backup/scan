// Package restful privides a function to start restful server
package restful

import "github.com/oniio/oniDNS/http/restful"

//start restful
func StartServer() {
	rt := restful.InitRestServer()
	go rt.Start()
}
