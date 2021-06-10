package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	rpc "github.com/hacash/service/server"
	"net/http"
	"strings"
)

/**
 * 按位置扫描一笔交易，并且找出里面的“转账操作”
 * 包括HAC、BTC和HACD的产生和转移
 */
func (api *RpcService) scanTransfersOfTransactionByPosition(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	height := CheckParamUint64(r, "height", 0) // 区块高度
	txposi := CheckParamUint64(r, "txposi", 0) // 交易索引位置

	txhash := CheckParamHex(r, "txhash", nil) // 交易索引位置
	if txhash != nil {
		if len(txhash) != 32 {
			ResponseErrorString(w, "param 'txhash' error")
			return
		}
	}

	// 是否以枚为单位
	isUnitMei := CheckParamBool(r, "unitmei", false)

	// kind = hsd
	kindStr := strings.ToLower(CheckParamString(r, "kind", ""))
	actAllKinds := false // 支持全部种类
	actKindHacash := false
	actKindSatoshi := false
	actKindDiamond := false
	if len(kindStr) == 0 {
		actAllKinds = true
	} else {
		if strings.Contains(kindStr, "h") {
			actKindHacash = true
		}
		if strings.Contains(kindStr, "s") {
			actKindSatoshi = true
		}
		if strings.Contains(kindStr, "d") {
			actKindDiamond = true
		}
	}

	// read tx
	var tx interfaces.Transaction = nil
	if height > 0 {
		blockObj, e := rpc.LoadBlockWithCache(api.backend.BlockChain().State(), height)
		if e != nil {
			ResponseError(w, e)
			return
		}
		blktxnum := blockObj.GetTransactionCount()
		txPosMargin := blktxnum - blockObj.GetCustomerTransactionCount()
		blktrs := blockObj.GetTransactions()
		realtxpos := uint32(txposi) + txPosMargin
		if realtxpos >= blktxnum || realtxpos >= uint32(len(blktrs)) {
			ResponseError(w, fmt.Errorf(" txposi <%d> overflow", txposi))
			return
		}
		// tx ok
		tx = blktrs[realtxpos]
		txhash = tx.Hash()
	} else if txhash != nil {
		// read tx body from disk
		blockheight, txbody, e := api.backend.BlockChain().State().BlockStore().ReadTransactionBytesByHash(txhash)
		if e != nil {
			ResponseError(w, e)
			return
		}
		if txbody == nil || len(txbody) == 0 {
			ResponseErrorString(w, "tx not find")
			return
		}
		txObj, _, e2 := transactions.ParseTransaction(txbody, 0)
		if e2 != nil {
			ResponseError(w, e2)
			return
		}
		tx = txObj
		height = blockheight
	} else {
		ResponseErrorString(w, "params error: height, txposi or txhash must give")
		return
	}

	// ret data
	var retdata = ResponseCreateData("type", tx.Type())
	trsActions := tx.GetActions()
	txfeepay := tx.GetFee()
	txfeegot := tx.GetFeeOfMinerRealReceived()
	retdata["hash"] = hex.EncodeToString(txhash)
	retdata["feepay"] = txfeepay.ToMeiOrFinString(isUnitMei)
	retdata["feegot"] = txfeegot.ToMeiOrFinString(isUnitMei)
	retdata["address"] = tx.GetAddress().ToReadable()
	retdata["height"] = height // block height
	retdata["timestamp"] = tx.GetTimestamp()

	effectiveActions := make([]interface{}, 0)
	// scan tx
	for i, act := range trsActions {
		var item = make(map[string]interface{})

		if tarAct, ok := act.(*actions.Action_1_SimpleToTransfer); ok && (actAllKinds || actKindHacash) {

			item["to"] = tarAct.ToAddress.ToReadable()
			item["hacash"] = tarAct.Amount.ToMeiOrFinString(isUnitMei)

		} else if tarAct, ok := act.(*actions.Action_13_FromTransfer); ok && (actAllKinds || actKindHacash) {

			item["from"] = tarAct.FromAddress.ToReadable()
			item["hacash"] = tarAct.Amount.ToMeiOrFinString(isUnitMei)

		} else if tarAct, ok := act.(*actions.Action_14_FromToTransfer); ok && (actAllKinds || actKindHacash) {

			item["from"] = tarAct.FromAddress.ToReadable()
			item["to"] = tarAct.ToAddress.ToReadable()
			item["hacash"] = tarAct.Amount.ToMeiOrFinString(isUnitMei)

		} else if tarAct, ok := act.(*actions.Action_7_SatoshiGenesis); ok && (actAllKinds || actKindSatoshi) {

			item["btctrsno"] = tarAct.TransferNo
			item["owner"] = tarAct.OriginAddress.ToReadable()
			item["satoshi"] = tarAct.BitcoinQuantity * 10000 * 10000 // unit: 1BTC = 1w * satoshi

		} else if tarAct, ok := act.(*actions.Action_8_SimpleSatoshiTransfer); ok && (actAllKinds || actKindSatoshi) {

			item["to"] = tarAct.ToAddress.ToReadable()
			item["satoshi"] = tarAct.Amount

		} else if tarAct, ok := act.(*actions.Action_4_DiamondCreate); ok && (actAllKinds || actKindDiamond) {

			item["number"] = tarAct.Number
			item["miner"] = tarAct.Address.ToReadable()
			item["diamond"] = string(tarAct.Diamond)

		} else if tarAct, ok := act.(*actions.Action_5_DiamondTransfer); ok && (actAllKinds || actKindDiamond) {

			item["to"] = tarAct.ToAddress.ToReadable()
			item["diamonds"] = string(tarAct.Diamond)

		} else if tarAct, ok := act.(*actions.Action_6_OutfeeQuantityDiamondTransfer); ok && (actAllKinds || actKindDiamond) {

			item["from"] = tarAct.FromAddress.ToReadable()
			item["to"] = tarAct.ToAddress.ToReadable()
			item["diamonds"] = tarAct.DiamondList.SerializeHACDlistToCommaSplitString()

		} else {
			continue
		}

		// ok
		item["ai"] = i
		effectiveActions = append(effectiveActions, item)
	}

	retdata["effective_actions"] = effectiveActions

	// return
	ResponseData(w, retdata)
	return
}
