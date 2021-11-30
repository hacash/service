package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/interfacev2"
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

	state := api.backend.BlockChain().StateRead()

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
	actKindChannel := false // 支持通道链
	actKindLending := false // 支持借贷
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
		if strings.Contains(kindStr, "l") {
			actKindLending = true
		}
		if strings.Contains(kindStr, "c") {
			actKindChannel = true
		}
	}

	// 借贷标记
	kindDiamondLending := (actAllKinds || (actKindDiamond && actKindLending))
	kindSatoshiLending := (actAllKinds || (actKindSatoshi && actKindLending))
	kindHacashLending := (actAllKinds || (actKindHacash && actKindLending))

	// read tx
	var tx interfacev2.Transaction = nil
	if height > 0 {
		blockObj, e := rpc.LoadBlockWithCache(state, height)
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
		blockheight, txbody, e := state.ReadTransactionBytesByHash(txhash)
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
		height = uint64(blockheight)
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

			// 通道链相关
		} else if _, ok := act.(*actions.Action_2_OpenPaymentChannel); ok && (actKindChannel) {

			// TODO::

			// 借贷相关
		} else if tarAct, ok := act.(*actions.Action_19_UsersLendingCreate); ok && (kindDiamondLending || kindSatoshiLending || kindHacashLending) {

			// 用户借贷抵押
			item["mortgagor"] = tarAct.MortgagorAddress.ToReadable() // 抵押人
			if kindDiamondLending && tarAct.MortgageDiamondList.Count > 0 {
				item["diamonds"] = tarAct.MortgageDiamondList.SerializeHACDlistToCommaSplitString()
			}
			if kindSatoshiLending && tarAct.MortgageBitcoin.NotEmpty.Check() {
				item["satoshi"] = tarAct.MortgageBitcoin.ValueSAT
			}
			if kindHacashLending {
				item["lender"] = tarAct.LenderAddress.ToReadable()                           // 债权人
				item["charge"] = tarAct.PreBurningInterestAmount.ToMeiOrFinString(isUnitMei) // 系统销毁的1%利息
				item["hacash"] = tarAct.LoanTotalAmount.ToMeiOrFinString(isUnitMei)          // 借出的HAC
			}

		} else if tarAct, ok := act.(*actions.Action_20_UsersLendingRansom); ok && (kindDiamondLending || kindSatoshiLending || kindHacashLending) {

			// 用户借贷抵押赎回
			item["redeemer"] = tx.GetAddress().ToReadable() // 赎回人 or 扣押人
			// 查询对象
			ldobj, _ := state.UserLending(tarAct.LendingID)
			if ldobj == nil {
				ResponseError(w, fmt.Errorf("User lending <%s> not find.", tarAct.LendingID.ToHex()))
				return
			}
			if kindDiamondLending && ldobj.MortgageDiamondList.Count > 0 {
				item["diamonds"] = ldobj.MortgageDiamondList.SerializeHACDlistToCommaSplitString()
			}
			if kindSatoshiLending && ldobj.MortgageBitcoin.NotEmpty.Check() {
				item["satoshi"] = ldobj.MortgageBitcoin.ValueSAT
			}
			if kindHacashLending {
				item["lender"] = ldobj.LenderAddress.ToReadable()                // 债权人
				item["hacash"] = tarAct.RansomAmount.ToMeiOrFinString(isUnitMei) // 归还的HAC（在劝人自己扣押则为零）
			}

		} else if tarAct, ok := act.(*actions.Action_15_DiamondsSystemLendingCreate); ok && (kindDiamondLending || kindHacashLending) {

			// 钻石系统借贷
			item["mortgagor"] = tx.GetAddress().ToReadable() // 抵押人
			if kindDiamondLending {                          // 抵押的 HACD
				item["diamonds"] = tarAct.MortgageDiamondList.SerializeHACDlistToCommaSplitString()
			}
			if kindHacashLending { // 从系统借出的 HAC
				item["hacash"] = tarAct.LoanTotalAmount.ToMeiOrFinString(isUnitMei)
			}

		} else if tarAct, ok := act.(*actions.Action_16_DiamondsSystemLendingRansom); ok && (kindDiamondLending || kindHacashLending) {

			// 钻石系统借贷赎回
			item["redeemer"] = tx.GetAddress().ToReadable() // 私有 or 公共赎回人
			// 查询对象
			ldobj, _ := state.DiamondSystemLending(tarAct.LendingID)
			if ldobj == nil {
				ResponseError(w, fmt.Errorf("Diamond system lending <%s> not find.", tarAct.LendingID.ToHex()))
				return
			}
			if kindDiamondLending { // 赎回的 HACD
				item["diamonds"] = ldobj.MortgageDiamondList.SerializeHACDlistToCommaSplitString()
			}
			if kindHacashLending { // 归还的 HAC
				item["hacash"] = tarAct.RansomAmount.ToMeiOrFinString(isUnitMei)
			}

		} else if tarAct, ok := act.(*actions.Action_17_BitcoinsSystemLendingCreate); ok && (kindSatoshiLending || kindHacashLending) {

			// 比特币系统借贷
			item["mortgagor"] = tx.GetAddress().ToReadable() // 抵押人
			if kindSatoshiLending {                          // 抵押的 HACD
				item["satoshi"] = uint64(tarAct.MortgageBitcoinPortion) * 100 * 10000 // 单位为0.01BTC
			}
			if kindHacashLending { // 从系统借出的 HAC
				item["hacash"] = tarAct.LoanTotalAmount.ToMeiOrFinString(isUnitMei)
				item["charge"] = tarAct.PreBurningInterestAmount.ToMeiOrFinString(isUnitMei) // 系统预先销毁的利息
			}

		} else if tarAct, ok := act.(*actions.Action_18_BitcoinsSystemLendingRansom); ok && (kindSatoshiLending || kindHacashLending) {

			// 钻石系统借贷赎回
			item["redeemer"] = tx.GetAddress().ToReadable() // 私有 or 公共赎回人
			// 查询对象
			ldobj, _ := state.BitcoinSystemLending(tarAct.LendingID)
			if ldobj == nil {
				ResponseError(w, fmt.Errorf("Bitcoin system lending <%s> not find.", tarAct.LendingID.ToHex()))
				return
			}
			if kindSatoshiLending { // 赎回的 HACD
				item["satoshi"] = uint64(ldobj.MortgageBitcoinPortion) * 100 * 10000 // 单位为0.01BTC
			}
			if kindHacashLending { // 归还的 HAC
				item["hacash"] = tarAct.RansomAmount.ToMeiOrFinString(isUnitMei)
			}

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
