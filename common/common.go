package common

import (
	"fmt"
	"net"
	"strings"

	"github.com/saveio/themis/common/log"
)

type ProtocolHostPort struct {
	Protocol string
	Host     string
	Port     string
}

func FullHostAddr(hostAddr, protocol string) string {
	if strings.Index(hostAddr, protocol) != -1 {
		return hostAddr
	}
	return fmt.Sprintf("%s://%s", protocol, hostAddr)
}

func FullHostPort(host, port string) string {
	if strings.Index(host, ":") != -1 {
		return host
	}
	return fmt.Sprintf("%s:%s", host, port)
}

func SplitHostAddr(endpointAddr string) (*ProtocolHostPort, error) {
	index := strings.Index(endpointAddr, "://")
	var hostPort, protocol string
	if index != -1 {
		hostPort = endpointAddr[index+3:]
		protocol = endpointAddr[0:index]
	}
	host, port, err := net.SplitHostPort(hostPort)
	log.Debugf("hostPort %v, host %v", hostPort, host)
	if err != nil {
		return nil, err
	}
	return &ProtocolHostPort{
		Protocol: protocol,
		Host:     host,
		Port:     port,
	}, nil
}
