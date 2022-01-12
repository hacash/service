package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	"strconv"
)

func (api *DeprecatedApiService) dexSellConfirm(params map[string]string) map[string]string {
	result := make(map[string]string)

	state := api.blockchain.GetChainEngineKernel().StateRead()
	//blockstore := state.BlockStoreRead()

	dmstr, ok1 := params["diamonds"]
	if !ok1 {
		result["err"] = "params diamonds must."
		return result
	}

	signstr, ok1 := params["signature"]
	if !ok1 {
		result["err"] = "params signature must."
		return result
	}
	signhex, e := hex.DecodeString(signstr)
	if e != nil {
		result["err"] = "params signature format error."
		return result
	}
	signck := fields.CreateSignCheckData("")
	_, e = signck.Parse(signhex, 0)
	if e != nil {
		result["err"] = "signature data error."
		return result
	}

	pricestr, ok3 := params["start_price"]
	if !ok3 {
		result["err"] = "params offer must."
		return result
	}
	start_price, e := fields.NewAmountFromString(pricestr)
	if e != nil {
		result["err"] = fmt.Sprintf("offer amount <%s> is error.", start_price)
		return result
	}

	diamonds := fields.NewEmptyDiamondListMaxLen200()
	e = diamonds.ParseHACDlistBySplitCommaFromString(dmstr)
	if e != nil {
		result["err"] = fmt.Sprintf("diamonds list error: ", e.Error())
		return result
	}

	// 检查卖方是否全部拥有钻石
	var seller fields.Address = nil
	for _, v := range diamonds.Diamonds {
		dia, e := state.Diamond(v)
		if e != nil {
			result["err"] = "read diamond state error."
			return result
		}
		if seller != nil && dia.Address.NotEqual(seller) {
			result["err"] = fmt.Sprintf("diamond <%s> not belong to seller address %s", v.Name(), seller.ToReadable())
			return result
		}
		seller = dia.Address
	}

	// 检查签名权限验证
	signok, addr, e := signck.VerifySign()
	if e != nil {
		result["err"] = "signature verify error."
		return result
	}
	if !signok {
		result["err"] = "signature verify fail."
		return result
	}
	if addr.NotEqual(seller) {
		result["err"] = fmt.Sprintf("signature verify fail, need address %s but got %s.", seller.ToReadable(), addr.ToReadable())
		return result
	}

	// 数据
	result["seller"] = seller.ToReadable()
	result["start_price"] = start_price.ToMeiString()
	result["diamond_count"] = strconv.Itoa(int(diamonds.Count))
	result["diamonds"] = diamonds.SerializeHACDlistToCommaSplitString()
	result["sign_data"] = signck.Stuffstr.Value()

	return result
}
