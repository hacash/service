package main

import (
	"bytes"
	"fmt"
	"github.com/buger/jsonparser"
	"strings"
)

func (r *Ranking) loadTotalSupply() error {

	resbts1, e1 := HttpGetBytes(r.node_rpc_url + "/query?action=total_supply")
	if e1 != nil {
		fmt.Println(e1)
		return fmt.Errorf("rpc not yet")
	}

	r.current_circulation, _ = jsonparser.GetFloat(resbts1, "current_circulation")
	r.minted_diamond, _ = jsonparser.GetInt(resbts1, "minted_diamond")
	r.transferred_bitcoin, _ = jsonparser.GetInt(resbts1, "transferred_bitcoin")

	fmt.Printf("load total supply %f, %d, %d.\n",
		r.current_circulation, r.minted_diamond, r.transferred_bitcoin)

	return nil
}

func (r *Ranking) changeDiamondsUnsafe(addrstr string, diamonds string, addOrSub bool) {
	alldiamonds := strings.Split(diamonds, ",")
	havdias := make(map[string]bool)
	for i := 0; i < len(alldiamonds); i++ {
		//fmt.Print(alldiamonds[i]+" ")
		havdias[alldiamonds[i]] = true
	}
	diatable, hav1 := r.cache_update_diamonds[addrstr]
	if !hav1 {
		// 从 磁盘加载
		v, e := r.ldb.Get([]byte("ds"+addrstr), nil)
		if e == nil {
			diatable = v // load ok
		}
	}
	//fmt.Print(" 0:", string(diatable))
	// 更新
	newdiabts := bytes.NewBuffer([]byte{})
	for i := 0; i+6 <= len(diatable); i += 6 {
		dian := diatable[i : i+6]
		_, ishav := havdias[string(dian)]
		if ishav {
			// 先删掉
			continue
		}
		// 本来的
		newdiabts.Write(dian)
	}
	//fmt.Print(" 1:", string(newdiabts.Bytes()))
	// 再添加回去
	if addOrSub {
		for i := 0; i < len(alldiamonds); i++ {
			newdiabts.Write([]byte(alldiamonds[i]))
		}
	}
	//fmt.Print(" 2:", string(newdiabts.Bytes()))
	// 更新
	//fmt.Println("r.cache_update_diamonds[addrstr] = newdiabts.Bytes()", addrstr, len(newdiabts.Bytes())/6)
	r.cache_update_diamonds[addrstr] = newdiabts.Bytes()
	// ok
}

// 添加一个待更新地址
func (r *Ranking) addWaitUpdateAddressUnsafe(addrstr string) {
	if len(addrstr) == 0 {
		return
	}
	if _, have := r.wait_update_address_list[addrstr]; have {
		return // 已经存在
	}
	r.wait_update_address_num += 1
	r.wait_update_address_list[addrstr] = true
	// 是否通知更新
	if r.wait_update_address_num == 50 {
		if len(r.flushStateToDiskNotifyCh) == 0 {
			go func() {
				r.flushStateToDiskNotifyCh <- true // 通知
			}()
		}
	}
}
