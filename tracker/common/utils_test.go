package common

import (
	"fmt"
	"testing"

	"github.com/saveio/themis/account"
	"github.com/saveio/themis/core/signature"
)

func Test_Sign(t *testing.T) {
	acc := account.NewAccount("AS8oaoWEmLpmv3CBvA228Ps7A89iM6WGSP")
	data := []byte{1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3}
	sig, err := signature.Sign(acc, data)
	fmt.Println(sig, len(sig), acc.PublicKey, err)

	err = signature.Verify(acc.PublicKey, data, sig)
	fmt.Println(err)
}
