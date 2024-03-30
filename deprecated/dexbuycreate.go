package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/transactions"
	"github.com/hacash/x16rs"
	"strconv"
	"strings"
)

func (api *DeprecatedApiService) dexBuyCreate(params map[string]string) map[string]string {
	result := make(map[string]string)

	state := api.blockchain.GetChainEngineKernel().StateRead()
	//blockstore := state.BlockStoreRead()

	dmstr, ok1 := params["diamonds"]
	if !ok1 {
		result["err"] = "params diamonds must."
		return result
	}

	buyerstr, ok2 := params["buyer"]
	if !ok2 {
		result["err"] = "params buyer must."
		return result
	}
	buyer, e := fields.CheckReadableAddress(buyerstr)
	if e != nil {
		result["err"] = fmt.Sprintf("buyer address <%s> is error.", buyerstr)
		return result
	}

	offerstr, ok3 := params["offer"]
	if !ok3 {
		result["err"] = "params offer must."
		return result
	}
	offer, e := fields.NewAmountFromString(offerstr)
	if e != nil {
		result["err"] = fmt.Sprintf("offer amount <%s> is error.", offerstr)
		return result
	}

	feeaddrstr, ok := params["fee_address"]
	if !ok {
		result["err"] = "params fee_address must."
		return result
	}
	feeaddr, e := fields.CheckReadableAddress(feeaddrstr)
	if e != nil {
		result["err"] = fmt.Sprintf("fee address <%s> is error.", feeaddrstr)
		return result
	}

	var fee = fields.NewEmptyAmount()
	feechargestr, ok := params["fee_charge"]
	if len(feechargestr) > 0 {
		// dex fee by set
		fee, e = fields.NewAmountFromString(feechargestr)
		if e != nil {
			result["err"] = fmt.Sprintf("fee charge amount <%s> is error.", feechargestr)
			return result
		}
	} else {
		// dex fee by ratio
		feeratiostr, ok := params["fee_ratio"]
		if !ok {
			result["err"] = "params fee_ratio must."
			return result
		}
		feeratio, e := strconv.ParseFloat(feeratiostr, 64)
		if e != nil {
			result["err"] = fmt.Sprintf("fee ratio <%s> is error.", feeratiostr)
			return result
		}

		if feeratio < 0 || feeratio >= 1 {
			result["err"] = fmt.Sprintf("fee ratio <%s> is error.", feeratiostr)
			return result
		}

		fee, e = fields.NewAmountFromMeiUnsafe(offer.ToMei() * feeratio)
		if e != nil {
			result["err"] = fmt.Sprintf("fee ratio <%s> is error.", feeratiostr)
			return result
		}
	}

	txfeestr, ok := params["tx_fee"]
	if !ok {
		result["err"] = "params tx_fee must."
		return result
	}

	tx_fee, e := fields.NewAmountFromString(txfeestr)
	if e != nil {
		result["err"] = fmt.Sprintf("tx_fee amount <%s> is error.", txfeestr)
		return result
	}

	dm200list := fields.NewEmptyDiamondListMaxLen200()
	e = dm200list.ParseHACDlistBySplitCommaFromString(dmstr)
	if e != nil {
		result["err"] = fmt.Sprintf("diamonds list error: ", e.Error())
		return result
	}

	dmslist := strings.Split(dmstr, ",")
	dmaddrs := make(map[string]int)
	var diamondaddr fields.Address = nil
	// Query diamond list
	for _, v := range dmslist {
		if false == x16rs.IsDiamondValueString(v) {
			result["err"] = fmt.Sprintf("<%s> is not a diamond name.", v)
			return result
		}
		dia, e := state.Diamond(fields.DiamondName(v))
		if e != nil {
			result["err"] = e.Error()
			return result
		}
		if dia == nil {
			result["err"] = fmt.Sprintf("<%s> not find.", v)
			return result
		}
		key := dia.Address.ToReadable()
		if _, ok := dmaddrs[key]; !ok {
			dmaddrs[key] = 0
		}
		diamondaddr = dia.Address
		dmaddrs[key] += 1
	}

	// All diamonds must belong to the same address
	if 1 != len(dmaddrs) {
		result["err"] = fmt.Sprintf("all diamonds must belong to one single address.")
		return result
	}

	// Create point-to-point transactions
	tx, e := transactions.NewEmptyTransaction_2_Simple(*buyer)
	if e != nil {
		result["err"] = fmt.Sprintf("tx create error: ", e.Error())
		return result
	}
	tx.SetFee(tx_fee)
	// Platform service charge
	if fee.IsNotEmpty() {
		tx.AddAction(&actions.Action_1_SimpleToTransfer{
			ToAddress: *feeaddr,
			Amount:    *fee,
		})
	}
	// Payment for purchase of drill
	tx.AddAction(&actions.Action_1_SimpleToTransfer{
		ToAddress: diamondaddr,
		Amount:    *offer,
	})
	// Get diamonds
	tx.AddAction(&actions.Action_6_OutfeeQuantityDiamondTransfer{
		FromAddress: diamondaddr,
		ToAddress:   *buyer,
		DiamondList: *dm200list,
	})

	// Return transaction content
	txbody, e := tx.Serialize()
	if e != nil {
		result["err"] = fmt.Sprintf("tx Serialize error: ", e.Error())
		return result
	}

	// ok
	result["tx_body"] = hex.EncodeToString(txbody)

	return result
}
