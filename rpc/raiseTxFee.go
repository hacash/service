package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/account"
	"github.com/hacash/core/fields"
	"net/http"
	"strings"
)

func (api *RpcService) raiseTxFee(r *http.Request, w http.ResponseWriter, bodybytes []byte) {
	var err error

	// mei
	isUnitMei := CheckParamBool(r, "unitmei", false)

	// address
	feePrikeyStr := strings.TrimLeft(CheckParamString(r, "fee_prikey", ""), "0x")
	feePrikeyStr = strings.TrimLeft(feePrikeyStr, "0X")
	if len(feePrikeyStr) == 0 {
		ResponseErrorString(w, "param fee_prikey must give")
		return
	}
	if len(feePrikeyStr) != 64 {
		ResponseErrorString(w, "param fee_prikey length error")
		return
	}
	var prikeybts []byte
	if prikeybts, err = hex.DecodeString(feePrikeyStr); err != nil {
		ResponseErrorString(w, "param fee_prikey format error")
		return
	}
	feeAccount, err := account.GetAccountByPriviteKey(prikeybts)

	// fee
	feeStr := CheckParamString(r, "fee", "")
	if len(feeStr) == 0 {
		ResponseErrorString(w, "param fee must give")
		return
	}
	var feeAmt *fields.Amount = nil
	if isUnitMei {
		feeAmt, err = fields.NewAmountFromMeiString(feeStr)
	} else {
		feeAmt, err = fields.NewAmountFromFinString(feeStr)
	}
	if err != nil {
		ResponseErrorString(w, "param fee format error")
		return
	}

	// txhash
	txhash := CheckParamHex(r, "txhash", nil) // 交易索引位置
	if len(txhash) != 32 {
		ResponseErrorString(w, "param 'txhash' error")
		return
	}

	// 查询交易
	tx, ok := api.txpool.CheckTxExistByHash(txhash)
	if !ok || tx == nil {
		ResponseErrorString(w, "Not find transaction in txpool.")
		return
	}

	// check
	if fields.Address(feeAccount.Address).Equal(tx.GetAddress()) != true {
		ResponseError(w, fmt.Errorf("Tx fee address password error: need %s but got %s", tx.GetAddress().ToReadable(), feeAccount.AddressReadable))
		return
	}

	// change fee
	tx = tx.Copy()
	tx.SetFee(feeAmt)

	// 私钥
	allPrivateKeyBytes := make(map[string][]byte, 1)
	allPrivateKeyBytes[string(feeAccount.Address)] = feeAccount.PrivateKey

	// do sign
	err3 := tx.FillNeedSigns(allPrivateKeyBytes, nil)
	if err3 != nil {
		ResponseError(w, err3)
		return
	}

	// add to pool
	err4 := api.txpool.AddTx(tx)
	if err4 != nil {
		ResponseError(w, err4)
		return
	}

	// return: status = success
	ResponseData(w, ResponseCreateData("status", "success"))

}
