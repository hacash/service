package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/account"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/transactions"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// Create a common transfer transaction
func (api *DeprecatedApiService) transferSimple(params map[string]string) map[string]string {
	result := make(map[string]string)
	password_or_privatekey, ok1 := params["password"]
	if !ok1 {
		result["err"] = "password must"
		return result
	}

	var acc *account.Account = nil
	privatekey, e2 := hex.DecodeString(password_or_privatekey)
	if len(password_or_privatekey) == 64 && e2 == nil {
		acc, e2 = account.GetAccountByPriviteKey(privatekey)
		if e2 != nil {
			result["err"] = "Privite Key Error"
			return result
		}
	} else {
		acc = account.CreateAccountByPassword(password_or_privatekey)
	}

	if strings.Compare(string(acc.AddressReadable), params["from"]) != 0 {
		result["err"] = "Privite Key error with address " + params["from"]
		return result
	}

	// Private key
	allPrivateKeyBytes := make(map[string][]byte, 1)
	allPrivateKeyBytes[string(acc.Address)] = acc.PrivateKey
	to_addr, e8 := fields.CheckReadableAddress(params["to"])
	if e8 != nil {
		result["err"] = e8.Error()
		return result
	}

	// Amount conversion
	amount, e6 := fields.NewAmountFromFinString(params["amount"])
	if e6 != nil {
		result["err"] = e6.Error()
		return result
	}
	fee, e7 := fields.NewAmountFromFinString(params["fee"])
	if e7 != nil {
		result["err"] = e7.Error()
		return result
	}

	// Create a general transfer transaction
	newTrs, e5 := transactions.NewEmptyTransaction_2_Simple(acc.Address)
	if e5 != nil {
		result["err"] = e5.Error()
		return result
	}
	curtime := time.Now()
	timestamp, ok := params["timestamp"]
	if ok {
		timebig, bb := new(big.Int).SetString(timestamp, 10)
		if !bb {
			result["err"] = "timestamp error"
			return result
		}
		curtime = time.Unix(timebig.Int64(), 0)
	}
	newTrs.Timestamp = fields.BlockTxTimestamp(curtime.Unix()) // Use current timestamp
	newTrs.Fee = *fee                                          // set fee
	tranact := actions.NewAction_1_SimpleToTransfer(*to_addr, amount)
	newTrs.AppendAction(tranact)

	// sign
	e9 := newTrs.FillNeedSigns(allPrivateKeyBytes, nil)
	if e9 != nil {
		result["err"] = e9.Error()
		return result
	}

	// Return transaction
	result["txhash"] = hex.EncodeToString(newTrs.HashFresh())
	txbody, e10 := newTrs.Serialize()
	if e10 != nil {
		result["err"] = e10.Error()
		return result
	}

	result["txbody"] = hex.EncodeToString(txbody)
	return result
}

// Query transaction confirmation status
func (api *DeprecatedApiService) txStatus(params map[string]string) map[string]string {
	result := make(map[string]string)
	result["status"] = "" // Indicates status

	// Transaction hash
	txhashstr, ok1 := params["txhash"]
	if !ok1 {
		result["err"] = "txhash must"
		return result
	}
	txhash, e2 := hex.DecodeString(txhashstr)
	if len(txhashstr) != 64 || e2 != nil {
		result["err"] = "txhash format error"
		return result
	}

	kernel := api.blockchain.GetChainEngineKernel()
	state := kernel.StateRead()
	store := state.BlockStoreRead()

	// Read from transaction pool
	_, ok2 := api.txpool.CheckTxExistByHash(txhash)
	if ok2 {
		// Transaction is in the trading pool
		result["status"] = "txpool" // Indicates that the transaction pool is in progress
		return result
	}

	// Query from block data
	blkhei, txbody, e3 := state.ReadTransactionBytesByHash(txhash)
	if e3 != nil {
		result["err"] = e3.Error()
		return result
	}

	if txbody == nil {
		result["status"] = "notfind" // Indicates that does not exist
		return result
	}

	// Exist and return
	result["status"] = "confirm" // Indicates that does not exist

	lastest, _, e4 := kernel.LatestBlock()
	if e4 != nil {
		result["err"] = e4.Error()
		return result
	}

	// Query block hash
	tarblkhash, e5 := store.ReadBlockHashByHeight(uint64(blkhei))
	if e5 != nil {
		result["err"] = e5.Error()
		return result
	}

	confirm_height := lastest.GetHeight() - uint64(blkhei)
	result["confirm_height"] = strconv.Itoa(int(confirm_height))    // Confirm the number of blocks
	result["block_height"] = strconv.FormatUint(uint64(blkhei), 10) // Block height
	result["block_hash"] = tarblkhash.ToHex()                       // Block hash

	return result
}
