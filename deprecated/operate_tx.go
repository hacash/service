package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/transactions"
	"net/http"
)

func (api *DeprecatedApiService) addTxToPool(w http.ResponseWriter, value []byte) {

	//fmt.Println(hex.EncodeToString(value))

	//defer func() {
	//	if err := recover(); err != nil {
	//		w.Write([]byte("Transaction body data error "))
	//	}
	//}()

	//fmt.Println("---- 1 ----")

	var tx, _, e = transactions.ParseTransaction(value, 0)
	if e != nil {
		w.Write([]byte("Transaction format error:\n" + e.Error()))
		return
	}
	//fmt.Println("---- 2 ----")
	//txbts, _ := tx.Serialize()
	//fmt.Println(hex.EncodeToString(txbts))
	//
	// 尝试加入交易池
	//fmt.Println("---- 3 ----")
	e3 := api.txpool.AddTx(tx)
	//fmt.Println("---- 4 ----")
	if e3 != nil {
		w.Write([]byte("Transaction Add to MemTxPool error: \n" + e3.Error()))
		return
	}

	//fmt.Println("---- 5 ----")
	// ok
	hashnofee := tx.Hash()
	hashnofeestr := hex.EncodeToString(hashnofee)
	w.Write([]byte("{\"success\":\"Transaction add to MemTxPool successfully !\",\"txhash\":\"" + hashnofeestr + "\"}"))

}
