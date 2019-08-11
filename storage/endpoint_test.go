package storage

import (
	"fmt"
	"net"
	"testing"
)

func TestPutEndpoint(t *testing.T) {
	dbPath := "../testdata/db1"
	db, err := NewLevelDBStore(dbPath)
	if err != nil || db == nil {
		return
	}
	epDB := NewEndpointDB(db)
	host := "127.0.0.1"
	port := 3333
	netHost := net.ParseIP(host)
	err = epDB.PutEndpoint("AUdg8DGDKvwUxyBbxpsMHSnGbYPxNCBSfg", netHost, int(port))
	fmt.Printf("put err:%s\n", err)
}

func TestGetEndpoint(t *testing.T) {
	dbPath := "../testdata/db1"
	db, err := NewLevelDBStore(dbPath)
	if err != nil || db == nil {
		return
	}
	epDB := NewEndpointDB(db)
	info, err := epDB.GetEndpoint("AUdg8DGDKvwUxyBbxpsMHSnGbYPxNCBSfg")
	fmt.Printf("info:%v, err:%s\n", info, err)
}

func TestDelEndpoint(t *testing.T) {
	dbPath := "../testdata/db1"
	db, err := NewLevelDBStore(dbPath)
	if err != nil || db == nil {
		return
	}
	epDB := NewEndpointDB(db)
	err = epDB.DelEndpoint("AUdg8DGDKvwUxyBbxpsMHSnGbYPxNCBSfg")
	fmt.Printf("put err:%s\n", err)
}
