package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/account"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"net/http"
	"strings"
)

/* test req:

http://127.0.0.1:8087/operate?action=raise_tx_fee&fee=2:244&fee_prikey=8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92&txhash=f7a9792c76de4ef9273c26510ec1b513f8a874ce34f6f41f3ab77386ff8ff7e7&txbody=0200656bf56500e63c33a796b3032ce6b856f68fccf06608d9ed18f401010001000100b776b849777097672459184230cfea3d6b6dc05bf8010100010231745adae24044ff09c3541537160abb8d5d720275bbaeed0b3d035b1e8b263cab5579cfb8d0e18998fa6b3ab5a1f6b34d2581ccff73a1dba96949e1e7e3f61d60af263fb8bcddb0748ee6cd87e5262318eb2e64bffbc4cfc9905b3e0285a5810000

http://127.0.0.1:8087//query?action=scan_value_transfers&txhash=f7a9792c76de4ef9273c26510ec1b513f8a874ce34f6f41f3ab77386ff8ff7e7

*/

func (api *RpcService) raiseTxFee(r *http.Request, w http.ResponseWriter, bodybytes []byte) {
	var err error

	// address
	feePrikeyStr := strings.TrimPrefix(CheckParamString(r, "fee_prikey", ""), "0x")
	feePrikeyStr = strings.TrimPrefix(feePrikeyStr, "0X")
	if len(feePrikeyStr) == 0 {
		ResponseErrorString(w, "param fee_prikey must give")
		return
	}

	// if feePrikeyStr size is not ok
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
	feeAmt, err = fields.NewAmountFromString(feeStr)
	if err != nil {
		ResponseErrorString(w, "param fee format error")
		return
	}

	// txhash
	txhash := CheckParamHex(r, "txhash", nil)
	if len(txhash) != 32 {
		ResponseErrorString(w, "param 'txhash' error")
		return
	}

	// Query transaction
	var tx interfaces.Transaction = nil
	var ok bool = false
	var txbody = CheckParamHex(r, "txbody", nil)
	if txbody != nil && len(txbody) > 0 {
		// read from txbody
		tx, _, err = transactions.ParseTransaction(txbody, 0)
		if tx == nil && err != nil {
			ResponseErrorString(w, fmt.Sprintf("ParseTransaction from body error: %s", err))
			return
		}
		var ttx = tx.Hash()
		if false == ttx.Equal(txhash) {
			ResponseErrorString(w, fmt.Sprintf("txhash %s not match txbody parse hash %s",
				hex.EncodeToString(txhash), hex.EncodeToString(ttx)))
			return
		}
	} else {
		// load from txpool
		tx, ok = api.txpool.CheckTxExistByHash(txhash)
		if !ok || tx == nil {
			ResponseErrorString(w, "Not find transaction in txpool.")
			return
		}
	}

	// check
	if fields.Address(feeAccount.Address).NotEqual(tx.GetAddress()) {
		ResponseError(w, fmt.Errorf("Tx fee address password error: need %s but got %s", tx.GetAddress().ToReadable(), feeAccount.AddressReadable))
		return
	}

	// change fee
	tx = tx.Clone()
	tx.SetFee(feeAmt)

	// Private key
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
