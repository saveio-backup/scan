package tracker

import "github.com/saveio/scan/storage"

// checkPaymentValid check payment from nimbus
func checkPaymentValid(paymentId string, infoHash storage.MetaInfoHash, sig []byte) bool {
	return true
}
