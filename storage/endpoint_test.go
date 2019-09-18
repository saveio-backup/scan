package storage

import (
	"fmt"
	"net"
	"testing"
)

var endpointDBPath string = "../testdata/db1"
var walletAddr string = "AUdg8DGDKvwUxyBbxpsMHSnGbYPxNCBSfg"

func TestPutEndpoint(t *testing.T) {
	db, err := NewLevelDBStore(endpointDBPath)
	if err != nil || db == nil {
		return
	}
	epDB := NewEndpointDB(db)
	host := "127.0.0.1"
	port := 3333
	netHost := net.ParseIP(host)
	err = epDB.PutEndpoint(walletAddr, netHost, int(port))
	fmt.Printf("put err:%s\n", err)
}

func TestGetEndpoint(t *testing.T) {
	db, err := NewLevelDBStore(endpointDBPath)
	if err != nil || db == nil {
		return
	}
	epDB := NewEndpointDB(db)
	info, err := epDB.GetEndpoint(walletAddr)
	fmt.Printf("info:%v, err:%s\n", info, err)
}

func TestDelEndpoint(t *testing.T) {
	db, err := NewLevelDBStore(endpointDBPath)
	if err != nil || db == nil {
		return
	}
	epDB := NewEndpointDB(db)
	err = epDB.DelEndpoint(walletAddr)
	fmt.Printf("put err:%s\n", err)
}
