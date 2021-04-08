package main

import (
	"fmt"
	"github.com/buger/jsonparser"
	"time"
)

// 扫描 rpc 接口
func (r *Ranking) scan() {
	time.Sleep(time.Second * 3) // 3秒后开始扫描
	// 循环扫描
	for {
		err := r.scanOneBlock()
		if err != nil {
			// 接口返回错误，等待十秒
			time.Sleep(time.Second * 10)
		}
	}
}

func (r *Ranking) scanOneBlock() error {
	r.dataChangeLocker.Lock()
	defer r.dataChangeLocker.Unlock()

	// 开始更新
	var scanHeight = r.finish_scan_block_height + 1
	if scanHeight%1000 == 0 {
		fmt.Printf("scan block %d.\n", scanHeight)
	}
	// fmt.Println("scanOneBlock:", scanHeight)
	blkUrl := fmt.Sprintf("/query?action=block_intro&unitmei=1&height=%d", scanHeight)
	resbts1, e1 := HttpGetBytes(r.node_rpc_url + blkUrl)
	if e1 != nil {
		// 接口未准备好
		return fmt.Errorf("rpc not yet")
	}
	// 获取交易数量
	txs, e2 := jsonparser.GetInt(resbts1, "transaction_count")
	if e2 != nil {
		return fmt.Errorf("rpc not yet")
	}
	rwdaddrstr, _ := jsonparser.GetString(resbts1, "coinbase", "address")
	if txs == 0 {
		// 空区块，没有交易
		r.addWaitUpdateAddressUnsafe(rwdaddrstr)
		// 标记本区块已经完成扫描
		r.finish_scan_block_height = scanHeight
		return nil // 成功返回
	}

	// 扫描交易
	for txposi := 0; txposi < int(txs); txposi++ {
		scanUrl := fmt.Sprintf("/query?action=scan_value_transfers&unitmei=1&height=%d&txposi=%d", scanHeight, txposi)
		resbts, e1 := HttpGetBytes(r.node_rpc_url + scanUrl)
		if e1 != nil {
			return fmt.Errorf("rpc not yet") // 错误
		}
		mainAddrStr, e3 := jsonparser.GetString(resbts, "address")
		if e3 != nil {
			return fmt.Errorf("rpc not yet") // 错误
		}
		r.addWaitUpdateAddressUnsafe(mainAddrStr) // 待更新地址
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
			v3, _ := jsonparser.GetString(value, "diamond")
			v4, _ := jsonparser.GetString(value, "diamonds")
			if len(v3) > 0 {
				v4 = v3
			}
			if len(v4) > 0 {
				// 写入钻石更新
				if len(v2) > 0 {
					r.changeDiamondsUnsafe(v2, v4, true) // 挖到
				} else {
					from := mainAddrStr
					if len(a1) > 0 {
						from = a1
					}
					r.changeDiamondsUnsafe(from, v4, false) // 转出
					to := mainAddrStr
					if len(a2) > 0 {
						to = a2
					}
					r.changeDiamondsUnsafe(to, v4, true) // 收到
				}
			}
		}, "effective_actions")
	}

	// 标记本区块已经完成扫描
	r.finish_scan_block_height = scanHeight

	// 成功返回
	return nil
}
