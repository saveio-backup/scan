package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"strings"

	"github.com/saveio/themis/common/log"
)

type ProtocolHostPort struct {
	Protocol string
	Host     string
	Port     string
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

type accountReader struct {
	PublicKey []byte
}

func (this accountReader) Read(buf []byte) (int, error) {
	bufs := make([]byte, 0)
	hash := sha256.Sum256(this.PublicKey)
	bufs = append(bufs, hash[:]...)
	log.Debugf("bufs :%s", hex.EncodeToString(bufs))
	for i, _ := range buf {
		if i < len(bufs) {
			buf[i] = bufs[i]
			continue
		}
		buf[i] = 0
	}
	return len(buf), nil
}
