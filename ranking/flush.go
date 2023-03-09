package main

import (
	"encoding/binary"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/hacash/core/fields"
	"strconv"
	"strings"
)

// Save data to disk
func (r *Ranking) flushStateToDisk() error {
	r.dataChangeLocker.Lock()
	defer r.dataChangeLocker.Unlock()

	tt1 := r.wait_update_address_num
	tt2 := len(r.cache_update_diamonds)

	// Query balance and update ranking table
	if r.wait_update_address_num > 0 {
		alladdrstrs := make([]string, 0)
		alladdrs := make([][]string, 0)
		cn := -1
		ci := -1

		for i, _ := range r.wait_update_address_list {
			cn++
			if cn%100 == 0 {
				ci += 1
				alladdrstrs = append(alladdrstrs, "")
				alladdrs = append(alladdrs, make([]string, 0))
			}
			alladdrstrs[ci] += "," + i
			alladdrs[ci] = append(alladdrs[ci], i)
		}

		// Interface request every 100 addresses
		for k := 0; k < len(alladdrstrs); k++ {
			addrstrs := strings.TrimLeft(alladdrstrs[k], ",")
			addrs := alladdrs[k]
			// Read RPC
			blsUrl := fmt.Sprintf("/query?action=balances&unitmei=1&address_list=%s", addrstrs)
			resbts1, e1 := HttpGetBytes(r.node_rpc_url + blsUrl)
			if e1 != nil {
				fmt.Println(e1)
				//os.Exit(0)
				// Interface not ready
				return fmt.Errorf("rpc not yet")
			}

			// Update balance table in sequence
			k1 := 0
			jsonparser.ArrayEach(resbts1, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				d, _ := jsonparser.GetInt(value, "diamond")
				s, _ := jsonparser.GetInt(value, "satoshi")
				hs, _ := jsonparser.GetString(value, "hacash")
				h, _ := strconv.ParseFloat(hs, 64)
				//fmt.Println(alladdrs[k1], d, s, h)
				item1 := NewBalanceRankingItem(addrs[k1], false)
				item1.SetBalanceByUint64(uint64(d))
				item2 := NewBalanceRankingItem(addrs[k1], false)
				item2.SetBalanceByUint64(uint64(s))
				item3 := NewBalanceRankingItem(addrs[k1], true)
				item3.SetBalanceByFloat64(float64(h))
				// to update
				r.diamond_balance_ranking_100 = UpdateBalanceRankingTable(r.diamond_balance_ranking_100, item1, r.balance_ranking_range)
				r.satoshi_balance_ranking_100 = UpdateBalanceRankingTable(r.satoshi_balance_ranking_100, item2, r.balance_ranking_range)
				r.hacash_balance_ranking_100 = UpdateBalanceRankingTable(r.hacash_balance_ranking_100, item3, r.balance_ranking_range)
				k1++
			}, "list")

			// Save balance table
			tb1 := SerializeBalanceRankingItems(r.hacash_balance_ranking_100)
			r.ldb.Put([]byte(DBKeyHacashBalanceRanking100), tb1, nil)
			tb2 := SerializeBalanceRankingItems(r.diamond_balance_ranking_100)
			r.ldb.Put([]byte(DBKeyDiamondBalanceRanking100), tb2, nil)
			tb3 := SerializeBalanceRankingItems(r.satoshi_balance_ranking_100)
			r.ldb.Put([]byte(DBKeySatoshiBalanceRanking100), tb3, nil)
		}
	}
	r.wait_update_address_list = make(map[string]bool)
	r.wait_update_address_num = 0

	// Update diamond table
	for i, v := range r.cache_update_diamonds {
		//fmt.Println("更新钻石表", i, string(v))
		r.ldb.Put([]byte("ds"+i), v, nil)
	}
	r.cache_update_diamonds = make(map[string][]byte, 0) // rebuild

	// Save scan block record
	hei := fields.BlockHeight(r.finish_scan_block_height)
	heibts, _ := hei.Serialize()
	r.ldb.Put([]byte(DBKeyFinishScanBlockHeight), heibts, nil)

	// Notification channel reset
	//r.flushStateToDiskNotifyCh

	// Print message
	fmt.Printf("flush height %d state %d, %d addresses , top max:", r.finish_scan_block_height, tt1, tt2)
	if len(r.hacash_balance_ranking_100) > 0 {
		it := r.hacash_balance_ranking_100[0]
		fmt.Printf(" HAC: %f(%d)", it.GetBalance(), len(r.hacash_balance_ranking_100))
	}

	if len(r.diamond_balance_ranking_100) > 0 {
		it := r.diamond_balance_ranking_100[0]
		fmt.Printf(" HACD: %d(%d)", it.GetBalanceForceUint64(), len(r.diamond_balance_ranking_100))
	}

	if len(r.satoshi_balance_ranking_100) > 0 {
		it := r.satoshi_balance_ranking_100[0]
		fmt.Printf(" SAT: %d(%d)", it.GetBalanceForceUint64(), len(r.satoshi_balance_ranking_100))
	}
	fmt.Printf(".\n")

	// success
	return nil
}

func (r *Ranking) flushTransferTurnover(count *TransferTurnoverStatistic) {
	var svkey = fmt.Sprintf("ttswk%d", count.WeekNum)
	var svdts = count.Serialize()
	r.ldb.Put([]byte(svkey), svdts, nil)
	// turnover_finish_block_height
	var svhei = make([]byte, 8)
	binary.BigEndian.PutUint64(svhei, r.turnover_finish_block_height)
	r.ldb.Put([]byte("tfbhei"), svhei, nil)
}

func (r *Ranking) loadTransferTurnoverFromDisk(weeknum uint32) *TransferTurnoverStatistic {
	var newturn = NewTransferTurnoverStatistic()
	if weeknum == 0 {
		return newturn
	}
	var svkey = fmt.Sprintf("ttswk%d", weeknum)
	valdts, _ := r.ldb.Get([]byte(svkey), nil)
	newturn.Parse(valdts, 0)
	// turnover_finish_block_height
	if r.turnover_finish_block_height <= 0 {
		var numdts, _ = r.ldb.Get([]byte("tfbhei"), nil)
		if numdts != nil {
			r.turnover_finish_block_height = binary.BigEndian.Uint64(numdts)
		}
	}
	// ok
	return newturn

}
