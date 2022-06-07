package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/account"
	"net/http"
)

// Batch create account
func (api *RpcService) createAccounts(r *http.Request, w http.ResponseWriter, bodybytes []byte) {
	number := int(CheckParamUint64(r, "number", 1))
	if number < 1 {
		number = 1
	}

	if number > 200 {
		number = 200
	}

	// create
	var lists = make([]interface{}, number)
	for i := 0; i < number; i++ {
		item := make(map[string]interface{})
		acc := account.CreateNewRandomAccount()
		item["prikey"] = hex.EncodeToString(acc.PrivateKey)
		item["pubkey"] = hex.EncodeToString(acc.PublicKey)
		item["address"] = acc.AddressReadable
		lists[i] = item
	}

	// return
	ResponseList(w, lists)
	return
}
