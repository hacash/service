package main

import (
	"fmt"
	"github.com/hacash/core/fields"
)

func (r *Ranking) init() error {

	// 供应量
	e1 := r.loadTotalSupply()
	if e1 != nil {
		fmt.Println(e1)
	}

	// 完成扫描的区块高度
	v, e := r.ldb.Get([]byte(DBKeyFinishScanBlockHeight), nil)
	//fmt.Println(v, e)
	if e == nil && len(v) == 5 {
		vb := fields.BlockHeight(0)
		if _, e2 := vb.Parse(v, 0); e2 == nil {
			r.finish_scan_block_height = uint64(vb)
			fmt.Printf("[Init] load finish_scan_block_height = %d.\n", vb)
		}
	}

	// 三张余额表
	v1, e1 := r.ldb.Get([]byte(DBKeyHacashBalanceRanking100), nil)
	if e1 == nil && len(v1) >= 21 {
		list1 := ParseBalanceRankingItems(v1)
		r.hacash_balance_ranking_100 = list1
	}
	v2, e2 := r.ldb.Get([]byte(DBKeyDiamondBalanceRanking100), nil)
	if e2 == nil && len(v2) >= 21 {
		list2 := ParseBalanceRankingItems(v2)
		r.diamond_balance_ranking_100 = list2
	}
	v3, e3 := r.ldb.Get([]byte(DBKeySatoshiBalanceRanking100), nil)
	if e3 == nil && len(v3) >= 21 {
		list3 := ParseBalanceRankingItems(v3)
		r.satoshi_balance_ranking_100 = list3
	}

	// ok
	return nil
}
