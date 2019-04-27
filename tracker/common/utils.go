/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-27
 */
package common

import (
	"encoding/json"
	"github.com/oniio/oniChain/common"
	"github.com/oniio/oniChain/common/log"
	"github.com/oniio/oniChain/core/signature"
	"github.com/oniio/oniChain/core/types"
	"github.com/oniio/oniChain/crypto/keypair"
	"github.com/oniio/oniChain/errors"
	"net"
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
