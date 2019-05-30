package common

import (
	"encoding/json"
	"net"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/signature"
	"github.com/saveio/themis/core/types"
	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/errors"
)

var payload ApiParams

type ApiParams struct {
	TrackerUrl string
	Wallet     [20]byte
	IP         net.IP
	Port       uint16
}

func ParamVerify(trackerUrl string, sigData []byte, pubKey keypair.PublicKey, walletAddr [20]byte, nodeIP net.IP, port uint16) error {
	address := types.AddressFromPubKey(pubKey)
	if address != common.Address(walletAddr) {
		return errors.NewErr("walletAddr and public key do not match!")
	}
	if nodeIP != nil && port != 0 {
		payload = ApiParams{
			TrackerUrl: trackerUrl,
			Wallet:     walletAddr,
			IP:         nodeIP,
			Port:       port,
		}
	} else {
		payload = ApiParams{
			TrackerUrl: trackerUrl,
			Wallet:     walletAddr,
		}
	}

	pb, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	err = signature.Verify(pubKey, pb, sigData)
	if err != nil {
		log.Errorf("signature verify error:%s", err)
		return err
	}
	return nil
}
