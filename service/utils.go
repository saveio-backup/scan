package service

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/saveio/themis/common/log"
)

const (
	MAX_DNS_TIMEWAIT_FOR_CHANNEL_CONNECT = time.Duration(10) * time.Second // max dns timewait for channel connect
	TRACKER_SERVICE_TIMEOUT              = time.Duration(15) * time.Second
	MAX_DNS_CHANNELS_NUM_AUTO_OPEN_WITH  = 5
)

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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func isValidUrl(dnsUrl string) bool {
	if strings.Index(dnsUrl, "0.0.0.0:0") != -1 {
		return false
	}
	return true
}
