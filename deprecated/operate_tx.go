package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"net/http"
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
func (api *DeprecatedApiService) CreateTxAndCheckOrCommit(w http.ResponseWriter, value []byte) {
	var e error = nil
	var readability = make([]string, 0, 4)
	var appendReadability = func(fmtstr string, a ...any) {
		readability = append(readability, fmt.Sprintf(fmtstr, a))
	}

	var returnError = func(e error) {
		//er := fmt.Errorf("")
		w.Write([]byte("{\"error\":\"Create transaction error: " + e.Error() + "\"}"))
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
	fee, e := jsonGetAmt(value, "fee")
	if e != nil {
		returnError(fmt.Errorf("Tx fee set error: %s", e.Error()))
		return
	}
	// addr
	madr, e := jsonGetAddr(value, "address")
	if e != nil {
		returnError(fmt.Errorf("Tx main address set error: %s", e.Error()))
		return
	}
	// timestamp
	tsnum, e := jsonparser.GetInt(value, "timestamp")
	if e != nil {
		tsnum = time.Now().Unix() // just now
	}
	// create tx
	tx, _ := transactions.NewEmptyTransaction_2_Simple(*madr)
	tx.Timestamp = fields.BlockTxTimestamp(uint64(tsnum))
	tx.Fee = *fee

	// block chain
	var kernel = api.backend.BlockChain().GetChainEngineKernel()

	var parse_act_err error = nil
	// parse actions
	jsonparser.ArrayEach(value, func(action []byte, dataType jsonparser.ValueType, offset int, err error) {
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
			toadr, e := jsonGetAddr(value, "to")
			if e != nil {
				parse_act_err = fmt.Errorf("Action address parse error: %s", e.Error())
				return
			}
			amt, e := jsonGetAmt(value, "amount")
			if e != nil {
				parse_act_err = fmt.Errorf("Action amount parse error: %s", e.Error())
				return
			}
			// tx append
			act := actions.NewAction_1_SimpleToTransfer(*toadr, amt)
			tx.AppendAction(act)
			// readability
			appendReadability("Transfer %sHAC to %s", amt.ToMeiString(), toadr.ToReadable())
			// ok
		} else if kind == 32 {
			// HACD inscription
			var hacdstr, e = jsonparser.GetString(action, "diamonds")
			var diamonds = fields.NewEmptyDiamondListMaxLen200()
			e = diamonds.ParseHACDlistBySplitCommaFromString(hacdstr)
			if e != nil {
				parse_act_err = fmt.Errorf("Action HACD name list parse error: %s", e.Error())
				return
			}
			// cost
			state := kernel.StateRead()
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
			tx.AppendAction(insact)
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
		returnError(fmt.Errorf("Tx actions parse error: %s", parse_act_err.Error()))
		return
	}

	ckstate, e := kernel.CurrentState().ForkSubChild()
	if e != nil {
		returnError(fmt.Errorf("Check Tx ForkSubChild error: %s", e.Error()))
		return
	}
	defer ckstate.Destory()

	// check tx with out signature
	e = tx.WriteInChainState(ckstate)
	if e != nil {
		returnError(fmt.Errorf("Check Tx WriteInChainState error: %s", e.Error()))
		return
	}

	// commit tx to blockchain if have signature of main address
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
			// Try to join the trading pool
			e = api.txpool.AddTx(tx)
			if e != nil {
				returnError(fmt.Errorf("Add Tx to txpool error: %s", e.Error()))
				return
			}
		}
	}

	// ok
	hashnofee := tx.Hash()
	txbody, _ := tx.Serialize()
	w.Write([]byte("{\"success\":\"ok\",\"txhash\":\"" + hashnofee.ToHex() +
		"\",\"txbody\":\"" + hex.EncodeToString(txbody) + "\"}"))

}
