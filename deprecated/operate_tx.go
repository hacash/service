package rpc

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"github.com/hacash/service/util/jsonparser"
	"net/http"
	"strings"
	"time"
)

func (api *DeprecatedApiService) addTxToPool(w http.ResponseWriter, value []byte) {
	var tx, _, e = transactions.ParseTransaction(value, 0)
	if e != nil {
		w.Write([]byte("Transaction format error:\n" + e.Error()))
		return
	}

	// Try to join the trading pool
	e3 := api.txpool.AddTx(tx.(interfaces.Transaction))
	if e3 != nil {
		w.Write([]byte("Transaction Add to MemTxPool error: \n" + e3.Error()))
		return
	}

	// ok
	hashnofee := tx.Hash()
	hashnofeestr := hex.EncodeToString(hashnofee)
	w.Write([]byte("{\"success\":\"Transaction add to MemTxPool successfully !\",\"txhash\":\"" + hashnofeestr + "\"}"))
}

/**
 * Create a transaction by JSON and attempt to validate it and commit to blockchain if give signature
 */
func (api *DeprecatedApiService) checkTx(req *http.Request, w http.ResponseWriter, value []byte) {

	//fmt.Println(string(value))
	var params = parseRequestQuery(req)

	sgadr, _ := params["sign_addr"]
	var sgaddr, _ = fields.CheckReadableAddress(sgadr)

	sgpubkey, _ := params["sign_pubkey"] // 33byte
	sgdata, _ := params["sign_data"]     // 64byte
	var sgpubkeyhex, _ = hex.DecodeString(sgpubkey)
	var sgdatahex, _ = hex.DecodeString(sgdata)

	// block chain
	var e error
	//var kernel = api.backend.BlockChain().GetChainEngineKernel()
	var tx interfaces.Transaction
	var readability []string

	// get from txbody
	tx, _, e = transactions.ParseTransaction(value, 0)

	// append sign
	if len(sgpubkeyhex) == 33 && len(sgdatahex) == 64 {
		sgs := tx.GetSigns()
		newsg := fields.Sign{
			PublicKey: sgpubkeyhex,
			Signature: sgdatahex,
		}
		var havesg = false
		for i := 0; i < len(sgs); i++ {
			if bytes.Equal(sgs[i].PublicKey, sgpubkeyhex) {
				sgs[i] = newsg // replase
				havesg = true
				break
			}
		}
		if !havesg {
			sgs = append(sgs, newsg) // append
		}
		tx.SetSigns(sgs) // reset
	}

	// parse from hex
	if e != nil {
		w.Write([]byte(fmt.Sprintf(`{"err":"Check transaction error: %s"}`, e.Error())))
		return
	}

	mainaddr := tx.GetAddress()
	readability = append(readability, fmt.Sprintf("Pay %sHAC tx fee by %s",
		tx.GetFee().ToMeiString(), mainaddr.ToReadable()))

	acts := tx.GetActionList()
	for i := 0; i < len(acts); i++ {
		var act = acts[i]
		if a, ok := act.(*actions.Action_1_SimpleToTransfer); ok {
			readability = append(readability, fmt.Sprintf("Transfer %sHAC from %s to %s",
				a.Amount.ToMeiString(), mainaddr.ToReadable(), a.ToAddress.ToReadable()))
		} else if a, ok := act.(*actions.Action_13_FromTransfer); ok {
			readability = append(readability, fmt.Sprintf("Transfer %sHAC from %s to %s",
				a.Amount.ToMeiString(), a.FromAddress.ToReadable(), mainaddr.ToReadable()))
		} else if a, ok := act.(*actions.Action_14_FromToTransfer); ok {
			readability = append(readability, fmt.Sprintf("Transfer %sHAC from %s to %s",
				a.Amount.ToMeiString(), a.FromAddress.ToReadable(), a.ToAddress.ToReadable()))
		} else if a, ok := act.(*actions.Action_5_DiamondTransfer); ok {
			readability = append(readability, fmt.Sprintf("Transfer 1 HACD (%s) from %s to %s",
				a.Diamond.Name(), mainaddr.ToReadable(), a.ToAddress.ToReadable()))
		} else if a, ok := act.(*actions.Action_6_OutfeeQuantityDiamondTransfer); ok {
			readability = append(readability, fmt.Sprintf("Transfer %d HACD (%s) from %s to %s",
				a.DiamondList.Count, a.DiamondList.SerializeHACDlistToCommaSplitString(), a.FromAddress.ToReadable(), a.ToAddress.ToReadable()))
		} else if a, ok := act.(*actions.Action_7_MultipleDiamondTransfer); ok {
			readability = append(readability, fmt.Sprintf("Transfer %d HACD (%s) from %s to %s",
				a.DiamondList.Count, a.DiamondList.SerializeHACDlistToCommaSplitString(), mainaddr.ToReadable(), a.ToAddress.ToReadable()))
		} else if a, ok := act.(*actions.Action_32_DiamondsEngraved); ok {
			readability = append(readability, fmt.Sprintf("Inscription '%s' to %d HACD (%s) by %s",
				a.EngravedContent.ShowString(), a.DiamondList.Count, a.DiamondList.SerializeHACDlistToCommaSplitString(), mainaddr.ToReadable()))
		} else {
			w.Write([]byte(fmt.Sprintf(`{"err":"Unsupported action kind: %d"}`, act.Kind())))
			return
		}

	}

	txbody, _ := tx.Serialize()
	// sign check
	needsigns, _ := tx.RequestSignAddresses(nil, false)
	sgaddrcks := []string{}
	for i := 0; i < len(needsigns); i++ {
		var addr = needsigns[i].ToReadable()
		var sgck = "false"
		if yes, _ := tx.VerifyTargetSigns([]fields.Address{needsigns[i]}); yes {
			sgck = "true"
		}
		sgaddrcks = append(sgaddrcks, fmt.Sprintf(`"%s":%s`, addr, sgck))
	}
	var sdadrckstr = strings.Join(sgaddrcks, ",")

	var sghx = ""
	if sgaddr != nil {
		if sgaddr.Equal(mainaddr) {
			sghx = tx.HashWithFee().ToHex()
		} else if strings.Contains(sdadrckstr, sgadr) {
			sghx = tx.Hash().ToHex()
		}
	}

	// ok
	w.Write([]byte(fmt.Sprintf(`{"sign_hash":"%s","hash":"%s","hash_with_fee":"%s","body":"%s","fee":"%s","address":"%s","timestamp":"%d",need_sign_address":{%s},"description":["%s"]}`,
		sghx, tx.Hash().ToHex(), tx.HashWithFee().ToHex(), hex.EncodeToString(txbody),
		tx.GetFee().ToMeiString(), mainaddr.ToReadable(), tx.GetTimestamp(),
		sdadrckstr, strings.Join(readability, `","`),
	)))

}

/**
 * Create a transaction by JSON and attempt to validate it and commit to blockchain if give signature
 */
func (api *DeprecatedApiService) createTxAndCheckOrCommit(w http.ResponseWriter, value []byte) {

	//fmt.Println(string(value))

	// block chain
	var e error
	var kernel = api.backend.BlockChain().GetChainEngineKernel()
	var tx interfaces.Transaction
	var readability []string

	// get from txbody
	txbdstr, e := jsonparser.GetString(value, "txbody")
	if e == nil && len(txbdstr) > 4 {
		txbts, e := hex.DecodeString(txbdstr)
		if e == nil {
			tx, _, _ = transactions.ParseTransaction(txbts, 0)
		}
	}

	// parse from json
	if tx == nil {
		tx, readability, e = createTransactionFromJson(kernel, value)
		if e != nil {
			w.Write([]byte(fmt.Sprintf(`{"err":"Create transaction error: %s"}`, e.Error())))
			return
		}
	}

	ckstate, e := kernel.CurrentState().ForkSubChild()
	if e != nil {
		w.Write([]byte(fmt.Sprintf(`{"err":"Check Tx ForkSubChild error: %s"}`, e.Error())))
		return
	}
	defer ckstate.Destory()

	// commit tx to blockchain if have signature of main address
	var isaddtopool = false
	pubkey_str, e1 := jsonparser.GetString(value, "pubkey")
	signature_str, e2 := jsonparser.GetString(value, "signature")
	if e1 == nil && e2 == nil {
		pubkey, e1 := hex.DecodeString(pubkey_str)
		signature, e2 := hex.DecodeString(signature_str)
		if e1 == nil && e2 == nil && len(pubkey) == 33 && len(signature) == 64 {
			tx.SetSigns([]fields.Sign{{
				PublicKey: pubkey,
				Signature: signature,
			}})
			/* txbody, _ := tx.Serialize()
			fmt.Println( hex.EncodeToString(tx.Hash()),
				hex.EncodeToString(txbody), len(txbody),
				tx.GetSigns()[0].GetAddress().ToReadable(),
				tx.GetSigns()[0].PublicKey.ToHex(),
				tx.GetSigns()[0].Signature.ToHex(),
				)*/
			// Try to join the trading pool
			e = api.txpool.AddTx(tx)
			if e != nil {
				e1 := fmt.Errorf("Add Tx to txpool error: %s", e.Error())
				//fmt.Println(e1.Error())
				w.Write([]byte(fmt.Sprintf(`{"success":false,"err":"%s"}`, e1.Error())))
				return
			}
			isaddtopool = true
		}
	}

	hashnofee := tx.Hash()
	txbody, _ := tx.Serialize()
	var txres = fmt.Sprintf(`,"txhash":"%s","txhashfee":"%s","txbody":"%s","txfee":"%s","description":["%s"]`,
		hashnofee.ToHex(),
		tx.HashWithFee().ToHex(),
		hex.EncodeToString(txbody),
		tx.GetFee().ToFinString(),
		strings.Join(readability, `","`),
	)

	if !isaddtopool {
		// check tx with out signature
		e = tx.WriteInChainState(ckstate)
		if e != nil {
			w.Write([]byte(fmt.Sprintf(`{"success":false%s,"err":"%s"}`, txres, e.Error())))
			return
		}

	}

	// ok
	w.Write([]byte(fmt.Sprintf(`{"success":true%s}`, txres)))

}

/**
 * create Transaction From Json
 */
func createTransactionFromJson(kernel interfaces.ChainEngine, jsonvalue []byte) (tx interfaces.Transaction, readability []string, err error) {
	state := kernel.StateRead()
	readability = make([]string, 0, 4)

	var appendReadability = func(fmtstr string, a ...any) {
		readability = append(readability, fmt.Sprintf(fmtstr, a...))
	}

	var jsonGetAmt = func(value []byte, key string) (*fields.Amount, error) {
		amtstr, e := jsonparser.GetString(value, key)
		if e != nil {
			return nil, fmt.Errorf("Not find amount '%s' in '%s'", key, string(value))
		}
		amt, e := fields.NewAmountFromString(amtstr)
		if e != nil {
			return nil, fmt.Errorf("Amount %s format error", amtstr)
		}
		// ok
		return amt, nil
	}
	var jsonGetAddr = func(value []byte, key string) (*fields.Address, error) {
		addrstr, e := jsonparser.GetString(value, key)
		if e != nil {
			return nil, fmt.Errorf("Not find address '%s' in '%s'", key, string(value))
		}
		addr, e := fields.CheckReadableAddress(addrstr)
		if e != nil {
			return nil, fmt.Errorf("Address %s format error", addrstr)
		}
		// ok
		return addr, nil
	}

	// fee
	fee, e := jsonGetAmt(jsonvalue, "fee")
	nopfee := false
	if e != nil || fee == nil || !fee.IsPositive() {
		// use recommon fee
		nopfee = true
		fee = fields.NewAmountByUnit(1, 244)
	}
	//fmt.Println("fee -------------- ", fee.ToFinString())
	// addr
	madr, e := jsonGetAddr(jsonvalue, "address")
	if e != nil {
		err = fmt.Errorf("Tx main address set error: %s", e.Error())
		return
	}
	// timestamp
	tsnum, e := jsonparser.GetInt(jsonvalue, "timestamp")
	if e != nil {
		tsnum = time.Now().Unix() // just now
	}
	// create tx
	txobj, e := transactions.NewEmptyTransaction_2_Simple(*madr)
	if e != nil {
		err = e
		return
	}
	txobj.Timestamp = fields.BlockTxTimestamp(uint64(tsnum))
	txobj.Fee = *fee
	tx = txobj

	// main address fee
	appendReadability(`%s as executed account and pay %sHAC tx fee`, madr.ToReadable(), fee.ToMeiString())

	var parse_act_err error = nil

	// parse actions
	jsonparser.ArrayEach(jsonvalue, func(action []byte, dataType jsonparser.ValueType, offset int, err error) {
		if parse_act_err != nil {
			return
		}

		var kind, e = jsonparser.GetInt(action, "kind")
		if e != nil {
			parse_act_err = fmt.Errorf("parse tx must set kind in '%s'", string(action))
			return
		}
		if kind == 1 {
			// HAC transfer
			toadr, e := jsonGetAddr(action, "to")
			if e != nil {
				parse_act_err = fmt.Errorf("Action address parse error: %s", e.Error())
				return
			}
			amt, e := jsonGetAmt(action, "amount")
			if e != nil {
				parse_act_err = fmt.Errorf("Action amount parse error: %s", e.Error())
				return
			}
			// tx append
			act := actions.NewAction_1_SimpleToTransfer(*toadr, amt)
			txobj.AddAction(act)
			// readability
			appendReadability("Transfer %sHAC to %s", amt.ToMeiString(), toadr.ToReadable())
			// ok
		} else if kind == 6 {
			// HACD transfer
			var hacdstr, e = jsonparser.GetString(action, "diamond")
			var diamonds = fields.NewEmptyDiamondListMaxLen200()
			e = diamonds.ParseHACDlistBySplitCommaFromString(hacdstr)
			if e != nil {
				parse_act_err = fmt.Errorf("Action HACD name list parse error: %s", e.Error())
				return
			}
			if diamonds.Count == 0 {
				parse_act_err = fmt.Errorf("Action HACD name empty")
				return
			}
			toadr, e := jsonGetAddr(action, "to")
			if e != nil {
				parse_act_err = fmt.Errorf("Action address parse error: %s", e.Error())
				return
			}
			var actobj interfaces.Action
			if diamonds.Count == 1 {
				actobj = &actions.Action_5_DiamondTransfer{
					Diamond:   diamonds.Diamonds[0],
					ToAddress: *toadr,
				}
			} else {
				actobj = &actions.Action_6_OutfeeQuantityDiamondTransfer{
					FromAddress: *madr,
					ToAddress:   *toadr,
					DiamondList: *diamonds,
				}
			}
			txobj.AddAction(actobj)
			// readability
			appendReadability("Transfer %dHACD (%s) to %s",
				diamonds.Count, diamonds.SerializeHACDlistToCommaSplitString(), toadr.ToReadable())

		} else if kind == 32 {
			// HACD inscription
			var hacdstr, e = jsonparser.GetString(action, "diamond")
			var diamonds = fields.NewEmptyDiamondListMaxLen200()
			e = diamonds.ParseHACDlistBySplitCommaFromString(hacdstr)
			if e != nil {
				parse_act_err = fmt.Errorf("Action HACD name list parse error: %s", e.Error())
				return
			}
			if diamonds.Count == 0 {
				parse_act_err = fmt.Errorf("Action HACD name empty")
				return
			}
			// cost
			//fmt.Println(state, diamonds, madr)///
			costamt, e := actions.RequestProtocolCostForDiamondList(state, diamonds, madr)
			if e != nil {
				parse_act_err = fmt.Errorf("Action check HACD belong or request protocol cost error: %s", e.Error())
				return
			}
			if costamt.Size() > 4 { // Compress size
				costamt, _, _ = costamt.CompressForMainNumLen(4, true)
			}
			// inscription
			insstr, e := jsonparser.GetString(action, "inscription")
			if e != nil {
				parse_act_err = fmt.Errorf("Action HACD inscription must set")
				return
			}
			insact, e := transactions.CreateOneActionOfHACDEngraved(diamonds, insstr, costamt)
			if e != nil {
				parse_act_err = fmt.Errorf("Action HACD inscription must set")
				return
			}
			//
			txobj.AddAction(insact)
			// readability
			readtip := fmt.Sprintf("Write HACD inscription '%s' for %s",
				insstr, diamonds.SerializeHACDlistToCommaSplitString())
			if costamt.IsNotEmpty() {
				readtip += fmt.Sprintf(" with pay(burn) %sHAC protocol fees")
			}
			readability = append(readability, readtip) // add tip
			// ok
		} else {
			parse_act_err = fmt.Errorf("parse tx action kind %d not support", kind)
		}
	}, "actions")

	if parse_act_err != nil {
		err = fmt.Errorf("Tx actions parse error: %s", parse_act_err.Error())
		return
	}

	if nopfee {

		addtxsize, _ := jsonparser.GetInt(jsonvalue, "addtxsize")
		feezhu := getLatestAverageFeePurityData(kernel, false, tx.Size()+2+uint32(addtxsize))
		fee = fields.NewAmountByUnit(int64(feezhu), 240)
		fee, _, e = fee.CompressForMainNumLen(2, true)
		if e != nil {
			err = fmt.Errorf("Tx Fee CompressForMainNumLen error: %s", e.Error())
			return
		}
		// reset fee
		txobj.Fee = *fee
		readability[0] = fmt.Sprintf(`%s as executed account and pay %sHAC tx fee`, madr.ToReadable(), fee.ToMeiString())
	}

	// ok
	return
}
