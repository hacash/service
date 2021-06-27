package main

import (
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/hacash/core/fields"
	"strconv"
	"strings"
)

// 保存数据到磁盘
func (r *Ranking) flushStateToDisk() error {
	r.dataChangeLocker.Lock()
	defer r.dataChangeLocker.Unlock()

	tt1 := r.wait_update_address_num
	tt2 := len(r.cache_update_diamonds)

	// 查询余额，更新排名表
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
		// 每100个地址请求一次接口
		for k := 0; k < len(alladdrstrs); k++ {
			addrstrs := strings.TrimLeft(alladdrstrs[k], ",")
			addrs := alladdrs[k]
			// 读取rpc
			blsUrl := fmt.Sprintf("/query?action=balances&unitmei=1&address_list=%s", addrstrs)
			resbts1, e1 := HttpGetBytes(r.node_rpc_url + blsUrl)
			if e1 != nil {
				fmt.Println(e1)
				//os.Exit(0)
				// 接口未准备好
				return fmt.Errorf("rpc not yet")
			}
			// 依次更新余额表
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
				// 更新
				r.diamond_balance_ranking_100 = UpdateBalanceRankingTable(r.diamond_balance_ranking_100, item1, r.balance_ranking_range)
				r.satoshi_balance_ranking_100 = UpdateBalanceRankingTable(r.satoshi_balance_ranking_100, item2, r.balance_ranking_range)
				r.hacash_balance_ranking_100 = UpdateBalanceRankingTable(r.hacash_balance_ranking_100, item3, r.balance_ranking_range)
				k1++
			}, "list")
			// 保存余额表
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

	// 更新钻石表
	for i, v := range r.cache_update_diamonds {
		//fmt.Println("更新钻石表", i, string(v))
		r.ldb.Put([]byte("ds"+i), v, nil)
	}
	r.cache_update_diamonds = make(map[string][]byte, 0) // 重设

	// 保存扫描区块记录
	hei := fields.BlockHeight(r.finish_scan_block_height)
	heibts, _ := hei.Serialize()
	r.ldb.Put([]byte(DBKeyFinishScanBlockHeight), heibts, nil)

	// 通知通道重设
	//r.flushStateToDiskNotifyCh

	// 打印消息
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

	// 成功
	return nil

}
