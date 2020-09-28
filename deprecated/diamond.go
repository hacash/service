package rpc

import (
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/stores"
	"github.com/hacash/core/transactions"
	"strconv"
	"strings"
)

//////////////////////////////////////////////////////////////

func (api *DeprecatedApiService) showDiamondCreateTxs(params map[string]string) map[string]string {
	result := make(map[string]string)

	txs := api.txpool.GetDiamondCreateTxs(-1)

	jsondata := []string{}
	for i, tx := range txs {
		for _, act := range tx.GetActions() {
			if dcact, ok := act.(*actions.Action_4_DiamondCreate); ok {
				fee := tx.GetFee()
				feeaddramt := api.blockchain.State().Balance(tx.GetAddress())
				status_code := 0 // ok
				if feeaddramt == nil || feeaddramt.Amount.LessThan(&fee) {
					status_code = 1 // 余额不足以支付手续费
				}
				jsondata = append(jsondata, fmt.Sprintf(`%d,"%s","%s","%s","%s","%s",%d`, i+1, tx.Hash().ToHex(), tx.GetAddress().ToReadable(),
					dcact.Diamond, dcact.Address.ToReadable(), fee.ToFinString(), status_code))
				break
			}
			if i >= 100 {
				break // max show num 100
			}
		}
	}
	perhei := 0
	lastest, _ := api.backend.BlockChain().State().ReadLastestBlockHeadAndMeta()
	if lastest != nil {
		perhei = (int(lastest.GetHeight()) + 5) / 5 * 5
	}
	liststr := strings.Join(jsondata, "],[")
	if len(liststr) > 0 {
		liststr = "[" + liststr + "]"
	}
	result["jsondata"] = `{"period":"` + strconv.Itoa(perhei) + `","datas":[` + liststr + `]}`
	return result
}

func (api *DeprecatedApiService) getDiamond(params map[string]string) map[string]string {
	result := make(map[string]string)
	dmstr, ok1 := params["name"]
	if !ok1 {
		result["err"] = "params name must."
		return result
	}

	state := api.blockchain.State()
	blockstore := state.BlockStore()

	var store *stores.DiamondSmelt = nil
	if dmnum, e := strconv.Atoi(dmstr); e == nil {
		store, _ = blockstore.ReadDiamondByNumber(uint32(dmnum))
	} else {
		if len(dmstr) != 6 {
			result["fail"] = "name format error."
			return result
		}
		store, _ = blockstore.ReadDiamond(fields.Bytes6(dmstr))
	}
	if store == nil {
		result["fail"] = "not find."
		return result
	}
	dmstr = string(store.Diamond)
	// get current belong
	sto2 := state.Diamond(fields.Bytes6(dmstr))
	if sto2 != nil {
		result["address"] = sto2.Address.ToReadable()
	} else {
		result["address"] = store.MinerAddress.ToReadable()
	}
	// ok
	//source_hash, _ := x16rs.Diamond(uint32(store.Number), store.PrevContainBlockHash, store.Nonce, store.MinerAddress, store.GetRealCustomMessage())
	result["name"] = dmstr
	result["current_block_hash"] = store.ContainBlockHash.ToHex()
	result["prev_block_hash"] = store.PrevContainBlockHash.ToHex()
	//result["source_hash"] = hex.EncodeToString(source_hash)
	result["block_height"] = strconv.FormatUint(uint64(store.ContainBlockHeight), 10)
	result["block_height"] = strconv.FormatUint(uint64(store.ContainBlockHeight), 10)
	result["number"] = strconv.Itoa(int(store.Number))
	result["miner_address"] = store.MinerAddress.ToReadable()
	result["custom_message"] = store.CustomMessage.ToHex()
	return result
}

func (api *DeprecatedApiService) transferDiamondMultiple(params map[string]string) map[string]string {

	result := make(map[string]string)

	feeAcc, err := api.readPasswordOrPriviteKeyParamBeAccount(params, "fee_password")
	if err != nil {
		result["err"] = err.Error()
		return result
	}

	diamondAcc, err2 := api.readPasswordOrPriviteKeyParamBeAccount(params, "diamond_password")
	if err2 != nil {
		result["err"] = err2.Error()
		return result
	}

	toAddress, err0 := fields.CheckReadableAddress(params["to_address"])
	if err0 != nil {
		result["err"] = err0.Error()
		return result
	}

	diamondstr, ok := params["diamonds"]
	if !ok {
		result["err"] = "param diamonds must"
		return result
	}
	diamonds := strings.Split(diamondstr, ",")
	if len(diamonds) > 200 {
		result["err"] = "too many diamond values"
		return result
	}

	// create tx
	tx, err3 := transactions.NewEmptyTransaction_2_Simple(feeAcc.Address)
	if err3 != nil {
		result["err"] = err3.Error()
		return result
	}
	feeBase := fields.NewAmountSmall(1, 244)
	feeAmount := feeBase.Copy()

	diamond_action := &actions.Action_6_OutfeeQuantityDiamondTransfer{}
	diamond_action.FromAddress = diamondAcc.Address
	diamond_action.ToAddress = *toAddress
	diamond_action.DiamondCount = fields.VarUint1(len(diamonds))
	diamond_action.Diamonds = make([]fields.Bytes6, len(diamonds))

	// append diamond action
	for i, v := range diamonds {
		if len(v) != 6 {
			result["err"] = v + " is not diamond value"
			return result
		}
		diamond_action.Diamonds[i] = fields.Bytes6(v)
		feeAmount, _ = feeAmount.Add(feeBase)
	}

	err4 := tx.AppendAction(diamond_action)
	if err4 != nil {
		result["err"] = err4.Error()
		return result
	}

	tx.Fee = *feeAmount

	// do sign
	allPrivateKeyBytes := make(map[string][]byte)
	allPrivateKeyBytes[string(feeAcc.Address)] = feeAcc.PrivateKey
	allPrivateKeyBytes[string(diamondAcc.Address)] = diamondAcc.PrivateKey
	//fmt.Println(allPrivateKeyBytes)

	e9 := tx.FillNeedSigns(allPrivateKeyBytes, nil)
	if e9 != nil {
		result["err"] = e9.Error()
		return result
	}

	// add to the tx pool
	err6 := api.txpool.AddTx(tx)
	if err6 != nil {
		result["err"] = err6.Error()
		return result
	}

	// ok
	result["status"] = "ok"
	return result

}
