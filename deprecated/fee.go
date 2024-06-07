package rpc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacashcom/core/account"
	"strconv"
	"strings"
)

func omitMoreDecimalForFloatString(number string) string {
	var pt = strings.IndexAny(number, ".")
	if pt == -1 {
		// not float with decimal
		return number
	}
	var n = len(number)
	var k = n
	for i := pt + 1; i < n; i++ {
		if number[i-1] != '0' && number[i] != '0' {
			k = i + 1
			break
		}
	}
	return number[0:k]
}

func (api *DeprecatedApiService) getLatestAverageFeePurity(params map[string]string) map[string]string {
	result := make(map[string]string)

	_, isunitmei := params["unitmei"]
	_, isomitdec := params["omitdec"]

	txszp, ok1 := params["txsize"]
	txsz := uint64(0)
	if ok1 {
		txsz, _ = strconv.ParseUint(txszp, 10, 64)
	}

	var kernel = api.blockchain.GetChainEngineKernel()
	//var feepb = uint64(kernel.GetLatestAverageFeePurity())
	//fmt.Println(feepb, "avgf feepb -------------")
	var feert = getLatestAverageFeePurityData(kernel, isunitmei, uint32(txsz))

	if txsz > 0 {
		var fpn string
		if isunitmei {
			fpn = fmt.Sprintf("%f", feert) // mei
		} else {
			fpn = strconv.FormatUint(uint64(feert), 10) // zhu
		}
		//fpn = "17"+fpn+"72656897"
		//fpn = "15.01"
		if isomitdec {
			fpn = omitMoreDecimalForFloatString(fpn)
		}
		result["feasible_fee"] = fpn

	} else {
		fpn := strconv.FormatUint(uint64(feert), 10)
		result["fee_purity"] = fpn
	}

	// ok
	return result

}

func getLatestAverageFeePurityData(kernel interfaces.ChainEngine, isunitmei bool, txsz uint32) float64 {

	var feepb = uint64(kernel.GetLatestAverageFeePurity())
	//fmt.Println(feepb, "avgf feepb -------------")

	if txsz > 0 {
		fp := feepb * (uint64(txsz)/32 + 1)
		if isunitmei {
			return float64(fp) / 10000_0000
			//fpn = fmt.Sprintf("%f", fpf)// mei
		} else {
			//fpn = strconv.FormatUint(fp, 10) // zhu
			return float64(fp)
		}
	} else {
		return float64(feepb)
	}
}

///////////////////////////////

// var aaaaaaaaaaaa = 0;
func (api *DeprecatedApiService) raiseTxFee(params map[string]string) map[string]string {
	result := make(map[string]string)
	_, isunitmei := params["unitmei"]
	/* // test
	if aaaaaaaaaaaa == 0 {
		aaaaaaaaaaaa++;
		txbts, _ := hex.DecodeString("0200666266ef00e63c33a796b3032ce6b856f68fccf06608d9ed18f401010001000100674e11e34c472ebfba2d34528fccd8aba826f2c4f8010100010231745adae24044ff09c3541537160abb8d5d720275bbaeed0b3d035b1e8b263cff1e092d81ac2df2c82ae50909d0222af6fc9695d9e2aabf68084cc8b6fb0ea77f3a2abc83ae77dde8cad006e287bfb30b8c5cf2eb98dea80bb478cfbb8aa1610000")
		tx1, _, _ := transactions.ParseTransaction(txbts, 0)
		fmt.Println( api.txpool.AddTx(tx1) )
		sg1 := tx1.GetSigns()[0]
		fmt.Println( hex.EncodeToString(sg1.PublicKey), hex.EncodeToString(sg1.Signature),  )
	} */
	// txhash
	txhashstr, _ := params["hash"]
	if len(txhashstr) != 64 {
		result["err"] = "params txhash format error."
		return result
	}
	txhash, _ := hex.DecodeString(txhashstr)
	if len(txhash) != 32 {
		result["err"] = "params txhash format error."
		return result
	}

	// load from txpool
	//fmt.Println(hex.EncodeToString(txhash))
	txpr, ok := api.txpool.CheckTxExistByHash(txhash)
	//fmt.Println(txpr, ok)
	if !ok || txpr == nil {
		result["err"] = "tx not find."
		return result
	}
	tx := txpr.Clone()

	feestr, _ := params["fee"]
	fee, _ := fields.NewAmountFromString(feestr)
	if fee == nil {
		result["err"] = "fee format error."
		return result
	}
	if !fee.MoreThan(tx.GetFee()) {
		result["err"] = fmt.Sprintf("reset fee must more than %s.",
			tx.GetFee().ToMeiOrFinString(isunitmei))
		return result
	}

	// ok change the fee
	tx.SetFee(fee)
	tx.ClearHash()
	hxwf := tx.HashWithFee()

	// if set sign
	isaddtxpool := false
	publickey, _ := hex.DecodeString(params["publickey"])
	signature, _ := hex.DecodeString(params["signature"])
	if len(publickey) == 33 && len(signature) == 64 {
		addr := account.NewAddressFromPublicKeyV0(publickey)
		if tx.GetAddress().NotEqual(addr) {
			result["err"] = fmt.Sprintf("tx main addr %s not match", tx.GetAddress().ToReadable())
			return result
		}
		signs := tx.GetSigns()
		for i := 0; i < len(signs); i++ {
			var li = signs[i]
			pubk, _ := li.PublicKey.Serialize()
			if bytes.Equal(pubk, publickey) {
				isaddtxpool = true
				signs[i] = fields.Sign{
					PublicKey: publickey,
					Signature: signature,
				}
				tx.SetSigns(signs) // update
				break
			}
		}
	}
	// add to the tx pool
	if isaddtxpool {
		ok, e := tx.VerifyAllNeedSigns()
		if !ok || e != nil {
			result["err"] = fmt.Sprintf("verify sign error: %s", e.Error())
			return result
		}
		e = api.txpool.AddTx(tx)
		//fmt.Println("-----------------",e)
		if e != nil {
			result["err"] = fmt.Sprintf("add tx to txpool error: %s", e.Error())
			return result
		}
	}

	// ok
	result["fee"] = tx.GetFee().ToMeiOrFinString(isunitmei)
	result["hash_with_fee"] = hex.EncodeToString(hxwf)
	result["hash"] = hex.EncodeToString(txhash)
	result["main_addr"] = tx.GetAddress().ToReadable()
	return result
}
