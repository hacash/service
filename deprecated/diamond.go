package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/stores"
	"github.com/hacash/core/transactions"
	"github.com/hacash/x16rs"
	"strconv"
	"strings"
)

//////////////////////////////////////////////////////////////

func (api *DeprecatedApiService) showDiamondCreateTxs(params map[string]string) map[string]string {
	result := make(map[string]string)

	_, isunitmei := params["unitmei"]
	txs := api.txpool.GetDiamondCreateTxs(-1)

	var kernel = api.backend.BlockChain().GetChainEngineKernel()

	var number int = 1
	diamd, _ := kernel.StateRead().ReadLastestDiamond()
	if diamd != nil {
		number = int(diamd.Number) + 1
	}

	jsondata := []string{}
	for i, tx := range txs {
		for _, act := range tx.GetActionList() {
			if dcact, ok := act.(*actions.Action_4_DiamondCreate); ok {
				fee := tx.GetFee()
				feeaddramt, _ := kernel.StateRead().Balance(tx.GetAddress())
				status_code := 0 // ok
				if feeaddramt == nil || feeaddramt.Hacash.LessThan(fee) {
					status_code = 1 // The balance is insufficient to pay the service charge
				}
				number = int(dcact.Number)
				jsondata = append(jsondata, fmt.Sprintf(`%d,"%s","%s","%s","%s","%s",%d`, i+1, tx.Hash().ToHex(), tx.GetAddress().ToReadable(),
					dcact.Diamond, dcact.Address.ToReadable(), fee.ToMeiOrFinString(isunitmei), status_code))
				break
			}
			if i >= 100 {
				break // max show num 100
			}
		}
	}

	perhei := 0
	acution_start_time := uint64(0)
	lastest, _, _ := kernel.LatestBlock()
	if lastest != nil {
		perhei = int(lastest.GetHeight()) / 5 * 5
		// get acution time
		if perhei > 0 {
			_, bdbts, _ := kernel.StateRead().BlockStoreRead().ReadBlockBytesByHeight(uint64(perhei))
			if bdbts != nil {
				blk, _, _ := blocks.ParseBlockHead(bdbts, 0)
				if blk != nil {
					acution_start_time = blk.GetTimestamp()
				}
			}
		}
		perhei += 5
	}

	liststr := strings.Join(jsondata, "],[")
	if len(liststr) > 0 {
		liststr = "[" + liststr + "]"
	}
	result["jsondata"] = `{"acution_start_time":` + strconv.FormatUint(acution_start_time, 10) + `,"period":` + strconv.Itoa(perhei) + `,"number":` + strconv.Itoa(number) + `,"datas":[` + liststr + `]}`

	return result
}

func (api *DeprecatedApiService) getDiamond(params map[string]string) map[string]string {
	result := make(map[string]string)
	_, isunitmei := params["unitmei"]
	dmstr, ok1 := params["name"]
	if !ok1 {
		result["err"] = "params name must."
		return result
	}

	// Late confirmation
	delayckn := uint32(0)
	dcnstr, ok2 := params["delay_confirm"]
	if ok2 {
		if dmnum, e := strconv.Atoi(dcnstr); e == nil {
			if dmnum > 0 {
				delayckn = uint32(dmnum)
			}
		}
	}

	state := api.blockchain.GetChainEngineKernel().StateRead()
	blockstore := state.BlockStoreRead()

	if delayckn > 0 {
		// Delay check
		if dmnum, e := strconv.Atoi(dmstr); e == nil {
			s1, _ := blockstore.ReadDiamondNameByNumber(uint32(dmnum) + delayckn)
			if len(s1) != fields.DiamondNameSize {
				// not find
				result["fail"] = "delay confirm check fail."
				return result
			}
		}
	}

	var store *stores.DiamondSmelt = nil
	if dmnum, e := strconv.Atoi(dmstr); e == nil {
		store, _ = blockstore.ReadDiamondByNumber(uint32(dmnum))
	} else {
		if len(dmstr) != 6 {
			result["fail"] = "name format error."
			return result
		}
		store, _ = blockstore.ReadDiamond(fields.DiamondName(dmstr))
	}

	if store == nil {
		result["fail"] = "not find."
		return result
	}

	dmstr = string(store.Diamond)
	// get current belong
	sto2, _ := state.Diamond(fields.DiamondName(dmstr))
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
	result["nonce"] = hex.EncodeToString(store.Nonce)
	result["block_height"] = strconv.FormatUint(uint64(store.ContainBlockHeight), 10)
	result["block_height"] = strconv.FormatUint(uint64(store.ContainBlockHeight), 10)
	result["number"] = strconv.Itoa(int(store.Number))
	result["miner_address"] = store.MinerAddress.ToReadable()
	result["custom_message"] = store.CustomMessage.ToHex()
	result["approx_fee_offer"] = store.GetApproxFeeOffer().ToMeiOrFinString(isunitmei)
	if isunitmei {
		result["average_burn_price"] = strconv.FormatUint(uint64(store.AverageBidBurnPrice), 10) + ".0"
	} else {
		result["average_burn_price"] = "ㄜ" + strconv.FormatUint(uint64(store.AverageBidBurnPrice), 10) + ":248"
	}
	result["visual_gene"] = store.VisualGene.ToHex()

	return result
}

func (api *DeprecatedApiService) getDiamondVisualGeneList(params map[string]string) map[string]string {
	result := make(map[string]string)

	// query
	state := api.blockchain.GetChainEngineKernel().StateRead()
	blockstore := state.BlockStoreRead()

	limit, _ := params["limit"]
	if len(limit) == 0 {
		limit = "10"
	}
	var limit_num uint64 = 10
	if lmt, err := strconv.ParseUint(limit, 10, 0); err == nil {
		limit_num = lmt
	}
	if limit_num > 50 {
		limit_num = 50 // Up to 50
	}

	jsondata := `{"list":[`
	var dtlist = make([]string, 0)

	dianamestr, ok0 := params["dianames"]
	start_number, ok1 := params["start_number"]
	isnewest, ok2 := params["newest"]
	_, isonly_gene_number := params["only_gene_number"]

	var appendOne = func(store *stores.DiamondSmelt) {
		if isonly_gene_number {
			dtlist = append(dtlist, fmt.Sprintf(
				`"%d,%s"`,
				store.Number,
				store.VisualGene.ToHex(),
			))
			return
		}
		dtlist = append(dtlist, fmt.Sprintf(
			`{"name":"%s","number":%d,"visual_gene":"%s","bid_fee":"%s"}`,
			string(store.Diamond),
			store.Number,
			store.VisualGene.ToHex(),
			store.ApproxFeeOffer.ToFinString()),
		)
	}

	if ok0 && len(dianamestr) >= 6 {

		// Query by name list
		diamonds := strings.Split(dianamestr, ",")
		dianum := 0
		for _, v := range diamonds {
			if x16rs.IsDiamondValueString(v) {
				store, e := blockstore.ReadDiamond(fields.DiamondName(v))
				if e != nil || store == nil {
					break
				}
				appendOne(store)
				dianum++
				if dianum >= 50 {
					break // Up to 50
				}
			}
		}

	} else if ok1 && len(start_number) > 0 {

		// Query by page
		start_number, ok1 = params["start_number"]
		if !ok1 {
			result["err"] = "params <start_number> must."
			return result
		}
		var start_num uint64 = 1
		if stt, err := strconv.ParseUint(start_number, 10, 0); err == nil {
			start_num = stt
		}

		for dianum := start_num; dianum < start_num+limit_num; dianum++ {

			store, e := blockstore.ReadDiamondByNumber(uint32(dianum))
			if e != nil || store == nil {
				break
			}
			appendOne(store)
		}

	} else if ok2 && len(isnewest) > 0 {

		lastdia, e := state.ReadLastestDiamond()
		if e != nil {
			result["err"] = "ReadLastestDiamond error."
			return result
		}
		diamonds := make([]*stores.DiamondSmelt, 0)
		diamonds = append(diamonds, lastdia)
		for dianum := lastdia.Number - 1; dianum >= 1 && dianum > lastdia.Number-fields.DiamondNumber(limit_num); dianum-- {
			store, e := blockstore.ReadDiamondByNumber(uint32(dianum))
			if e != nil || store == nil {
				break
			}
			diamonds = append(diamonds, store)
		}
		for _, dia := range diamonds {
			appendOne(dia)
		}

	}

	jsondata += strings.Join(dtlist, ",")
	jsondata += `]}`
	result["jsondata"] = jsondata

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

	var diamonds = fields.NewEmptyDiamondListMaxLen200()
	e11 := diamonds.ParseHACDlistBySplitCommaFromString(diamondstr)
	if e11 != nil {
		result["err"] = e11.Error()
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
	diamond_action.DiamondList = *diamonds

	// append diamond action
	for range diamonds.Diamonds {
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
