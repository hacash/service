package main

import (
	"fmt"
	"github.com/buger/jsonparser"
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

	// Mark that this block has been scanned
	r.finish_scan_block_height = scanHeight

	// Successful return
	return nil
}
