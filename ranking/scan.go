package main

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/hacash/core/fields"
	"strconv"
	"strings"
	"time"
)

// Scan RPC interface
func (r *Ranking) scan() {
	time.Sleep(time.Second * 3) // Start scanning in 3 seconds
	// Cyclic scanning
	for {
		err := r.scanOneBlock()
		if err != nil {
			// The interface returns an error. Wait for 10 seconds
			time.Sleep(time.Second * 10)
		}
	}
}

func (r *Ranking) scanOneBlock() error {
	r.dataChangeLocker.Lock()
	defer r.dataChangeLocker.Unlock()

	// Start update
	var scanHeight = r.finish_scan_block_height + 1
	if scanHeight%1000 == 0 {
		fmt.Printf("scan block %d.\n", scanHeight)
	}

	blkUrl := fmt.Sprintf("/query?action=block_intro&unitmei=1&height=%d", scanHeight)
	resbts1, e1 := HttpGetBytes(r.node_rpc_url + blkUrl)
	if e1 != nil {
		// Interface not ready
		return fmt.Errorf("rpc not yet")
	}

	// Get transaction quantity
	txs, e2 := jsonparser.GetInt(resbts1, "transaction_count")
	if e2 != nil {
		return fmt.Errorf("rpc not yet")
	}

	rwdaddrstr, _ := jsonparser.GetString(resbts1, "coinbase", "address")
	if txs == 0 {
		// Empty block, no transaction
		r.addWaitUpdateAddressUnsafe(rwdaddrstr)
		// Mark that this block has been scanned
		r.finish_scan_block_height = scanHeight
		return nil // Successful return
	}

	// Scan transactions
	for txposi := 0; txposi < int(txs); txposi++ {
		// Scan HAC transfer, BTC transfer, diamond mining, diamond transfer and diamond lending related actions
		scanUrl := fmt.Sprintf("/query?action=scan_value_transfers&unitmei=1&height=%d&txposi=%d&kind=hsdl", scanHeight, txposi)
		resbts, e1 := HttpGetBytes(r.node_rpc_url + scanUrl)
		if e1 != nil {
			return fmt.Errorf("rpc not yet") // error
		}

		mainAddrStr, e3 := jsonparser.GetString(resbts, "address")
		if e3 != nil {
			return fmt.Errorf("rpc not yet") // error
		}

		r.addWaitUpdateAddressUnsafe(mainAddrStr) // Address to be updated
		// actions
		jsonparser.ArrayEach(resbts, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			//*
			r.scanTurnoverFromEffectiveAction(scanHeight, value)
			//*
			a1, _ := jsonparser.GetString(value, "from")
			r.addWaitUpdateAddressUnsafe(a1)
			a2, _ := jsonparser.GetString(value, "to")
			r.addWaitUpdateAddressUnsafe(a2)
			v1, _ := jsonparser.GetString(value, "owner")
			r.addWaitUpdateAddressUnsafe(v1)
			v2, _ := jsonparser.GetString(value, "miner")
			r.addWaitUpdateAddressUnsafe(v2)
			l1, _ := jsonparser.GetString(value, "mortgagor") // 抵押人
			r.addWaitUpdateAddressUnsafe(l1)
			l2, _ := jsonparser.GetString(value, "redeemer") // 赎回人
			r.addWaitUpdateAddressUnsafe(l2)
			v3, _ := jsonparser.GetString(value, "diamond")
			v4, _ := jsonparser.GetString(value, "diamonds")
			if len(v3) > 0 {
				v4 = v3
			}

			if len(v4) > 0 {
				// Write diamond update
				if len(v2) > 0 {
					r.changeDiamondsUnsafe(v2, v4, true) // Dig to
				} else if len(l1) > 0 {
					r.changeDiamondsUnsafe(l1, v4, false) // mortgage
				} else if len(l2) > 0 {
					r.changeDiamondsUnsafe(l2, v4, true) // redeem
				} else {
					from := mainAddrStr
					if len(a1) > 0 {
						from = a1
					}
					r.changeDiamondsUnsafe(from, v4, false) // Transfer out
					to := mainAddrStr
					if len(a2) > 0 {
						to = a2
					}
					r.changeDiamondsUnsafe(to, v4, true) // received
				}
			}
		}, "effective_actions")
	}

	// mark data update
	r.cache_turnover_curobj.UpdateTime = time.Now()

	// Mark that this block has been scanned
	r.finish_scan_block_height = scanHeight

	// Successful return
	return nil
}

func (r *Ranking) scanTurnoverFromEffectiveAction(blkhei uint64, actionvalue []byte) {
	if blkhei <= r.turnover_finish_block_height {
		return // repeat
	}
	r.turnover_finish_block_height = blkhei
	var keyWeek = uint32(blkhei / 2000)
	var prevWeek = uint32(r.cache_turnover_curobj.WeekNum)
	if keyWeek != prevWeek {
		var newtts = NewTransferTurnoverStatistic()
		if prevWeek == 0 {
			newtts = r.loadTransferTurnoverFromDisk(keyWeek)
		}
		// save prev turnover
		//fmt.Printf("++++++++ %d\n", r.cache_turnover_curobj.WeekNum)
		go r.flushTransferTurnover(r.cache_turnover_curobj)
		newtts.WeekNum = fields.VarUint4(keyWeek)
		r.cache_turnover_curobj = newtts
	}
	// scan action
	from, _ := jsonparser.GetString(actionvalue, "from")
	to, _ := jsonparser.GetString(actionvalue, "to")
	var isdia_trs = len(from) > 0 || len(to) > 0

	hacash_str, _ := jsonparser.GetString(actionvalue, "hacash")
	var hacash, _ = strconv.ParseFloat(hacash_str, 64)
	if isdia_trs && hacash > 0 {
		r.cache_turnover_curobj.AppendHAC(hacash)
	}
	satoshi, _ := jsonparser.GetInt(actionvalue, "satoshi")
	if isdia_trs && satoshi > 0 {
		r.cache_turnover_curobj.AppendSAT(uint64(satoshi))
	}
	diamond, _ := jsonparser.GetString(actionvalue, "diamond")
	diamonds, _ := jsonparser.GetString(actionvalue, "diamonds")
	if len(diamond) == 6 {
		diamonds = diamond
	}
	if isdia_trs && len(diamonds) >= 6 {
		dianum := len(strings.Split(diamonds, ","))
		r.cache_turnover_curobj.AppendHACD(uint32(dianum))
	}
	// save settimeout

}
