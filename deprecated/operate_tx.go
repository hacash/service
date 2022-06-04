package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"net/http"
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
