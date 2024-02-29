package main

import (
	"bytes"
	"fmt"
	"../util/jsonparser"
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
		// Load from disk
		v, e := r.ldb.Get([]byte("ds"+addrstr), nil)
		if e == nil {
			diatable = v // load ok
		}
	}

	// to update
	newdiabts := bytes.NewBuffer([]byte{})
	for i := 0; i+6 <= len(diatable); i += 6 {
		dian := diatable[i : i+6]
		_, ishav := havdias[string(dian)]
		if ishav {
			// Delete first
			continue
		}
		// Original
		newdiabts.Write(dian)
	}

	// Add back
	if addOrSub {
		for i := 0; i < len(alldiamonds); i++ {
			newdiabts.Write([]byte(alldiamonds[i]))
		}
	}

	// to update
	r.cache_update_diamonds[addrstr] = newdiabts.Bytes()
	// ok
}

// Add an address to be updated
func (r *Ranking) addWaitUpdateAddressUnsafe(addrstr string) {
	if len(addrstr) == 0 {
		return
	}
	if _, have := r.wait_update_address_list[addrstr]; have {
		return // Already exists
	}
	r.wait_update_address_num += 1
	r.wait_update_address_list[addrstr] = true
	// Notify update
	if r.wait_update_address_num == 50 {
		if len(r.flushStateToDiskNotifyCh) == 0 {
			go func() {
				r.flushStateToDiskNotifyCh <- true // notice
			}()
		}
	}
}
