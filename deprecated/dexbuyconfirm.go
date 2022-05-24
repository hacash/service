package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/transactions"
	"strconv"
)

func (api *DeprecatedApiService) dexBuyConfirm(params map[string]string) map[string]string {
	result := make(map[string]string)

	state := api.blockchain.GetChainEngineKernel().StateRead()

	// 是否检查所有签名
	ckallsg, _ := params["check_all_sign"]
	ischeckallsign := len(ckallsg) > 0

	txbodystr, ok1 := params["tx_body_signed"]
	if !ok1 {
		result["err"] = "params tx_body_signed must."
		return result
	}

	txbody, e := hex.DecodeString(txbodystr)
	if e != nil {
		result["err"] = "tx_body_signed data error."
		return result
	}

	// parse tx
	tx, _, e := transactions.ParseTransaction(txbody, 0)
	if e != nil {
		result["err"] = "ParseTransaction error."
		return result
	}

	actionlist := tx.GetActionList()
	actlen := len(actionlist)
	if actlen != 2 && actlen != 3 {
		result["err"] = "Action length error."
		return result
	}

	// 检查余额
	minbalance := tx.GetFee().Copy()

	// 解析 tx
	var buyer = tx.GetAddress()
	var seller1 fields.Address = nil
	var seller2 fields.Address = nil
	var diamonds *fields.DiamondListMaxLen200 = nil
	var totalprice *fields.Amount = nil
	for _, v := range actionlist {
		if act, ok := v.(*actions.Action_1_SimpleToTransfer); ok {
			// 价格和卖家
			if totalprice == nil || act.Amount.MoreThan(totalprice) {
				totalprice = &act.Amount
				seller1 = act.ToAddress
			}
			minbalance, _ = minbalance.Add(&act.Amount)
		} else if act, ok := v.(*actions.Action_6_OutfeeQuantityDiamondTransfer); ok {
			if seller2 != nil {
				result["err"] = "action repeat error."
				return result
			}
			if act.ToAddress.NotEqual(buyer) {
				// 买卖家不匹配
				result["err"] = "buyer address error."
				return result
			}
			seller2 = act.FromAddress
			diamonds = &act.DiamondList
		} else {
			result["err"] = "Action type error."
			return result
		}
	}

	// 检查买卖方地址匹配
	if seller1 == nil || seller2 == nil || seller1.NotEqual(seller2) {
		result["err"] = "seller address not match."
		return result
	}

	// 检查买方签名
	signok, _ := tx.VerifyTargetSigns([]fields.Address{buyer})
	if signok == false {
		result["err"] = "Verify buyer sign fail."
		return result
	}

	// 是否检查所有签名
	if ischeckallsign {
		signok, _ := tx.VerifyAllNeedSigns()
		if signok == false {
			result["err"] = "Verify signatures fail."
			return result
		}
	}

	// 检查买方余额
	realbls, e := state.Balance(buyer)
	if e != nil || realbls == nil {
		result["err"] = "buyer get balance error."
		return result
	}
	if realbls.Hacash.LessThan(minbalance) {
		result["err"] = fmt.Sprintf("buyer %s balance not enough, need at least %s but got %s", buyer.ToReadable(), minbalance.ToFinString(), realbls.Hacash.ToFinString())
		return result
	}

	// 检查卖方是否全部拥有钻石
	for _, v := range diamonds.Diamonds {
		dia, e := state.Diamond(v)
		if e != nil {
			result["err"] = "read diamond state error."
			return result
		}
		if dia.Address.NotEqual(seller1) {
			result["err"] = fmt.Sprintf("diamond <%s> not belong to %s", v.Name(), seller1.ToReadable())
			return result
		}
	}

	// 数据
	result["tx_hash"] = tx.Hash().ToHex()
	result["buyer"] = buyer.ToReadable()
	result["seller"] = seller1.ToReadable()
	result["total_price"] = totalprice.ToMeiString()
	result["single_price"] = strconv.FormatFloat(totalprice.ToMei()/float64(diamonds.Count), 'g', 8, 64)
	result["diamond_count"] = strconv.Itoa(int(diamonds.Count))
	result["diamonds"] = diamonds.SerializeHACDlistToCommaSplitString()

	return result
}
