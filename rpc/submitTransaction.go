package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"net/http"
)

// Submit a transaction
func (api *RpcService) submitTransaction(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	isHexData := CheckParamBool(r, "hexbody", false)

	// Hex string mode
	if isHexData {
		realbodybts, e2 := hex.DecodeString(string(bodybytes))
		if e2 != nil {
			ResponseError(w, e2)
			return
		}
		//fmt.Println(len(realbodybts))
		//fmt.Println(string(realbodybts))
		bodybytes = realbodybts
	}

	// Parsing transactions
	trs, _, e3 := transactions.ParseTransaction(bodybytes, 0)
	if e3 != nil {
		ResponseError(w, e3)
		return
	}

	// Try to join the trading pool
	e4 := api.txpool.AddTx(trs.(interfaces.Transaction))
	if e4 != nil {
		ResponseError(w, e4)
		return
	}

	// Return success
	// return: status = success
	ResponseData(w, ResponseCreateData("status", "success"))

	return
}
