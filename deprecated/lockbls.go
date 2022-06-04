package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/stores"
	"math/big"
	"strconv"
)

// Query linear lock information
func (api *DeprecatedApiService) getLockBlsInfo(params map[string]string) map[string]string {
	result := make(map[string]string)
	lockbls_id, ok1 := params["lockbls_id"]
	if !ok1 {
		result["err"] = "param lockbls_id must."
		return result
	}

	lockbls_key, e1 := hex.DecodeString(lockbls_id)
	if e1 != nil {
		result["err"] = "param lockbls_id format error."
		return result
	}

	if len(lockbls_key) != stores.LockblsIdLength {
		result["err"] = "param lockbls_id length error."
		return result
	}

	// query
	state := api.blockchain.GetChainEngineKernel().StateRead()
	lockblsItem, _ := state.Lockbls(lockbls_key)
	if lockblsItem == nil {
		result["err"] = "not find."
		return result
	}

	// Return information
	result["master_address"] = lockblsItem.MasterAddress.ToReadable()
	result["effect_block_height"] = strconv.FormatUint(uint64(lockblsItem.EffectBlockHeight), 10)
	result["linear_block_number"] = strconv.FormatUint(uint64(lockblsItem.LinearBlockNumber), 10)
	amt1 := lockblsItem.TotalLockAmount
	result["total_lock_amount"] = amt1.ToFinString()
	amt2 := lockblsItem.LinearReleaseAmount
	result["linear_release_amount"] = amt2.ToFinString()
	amt3 := lockblsItem.BalanceAmount
	result["balance_amount"] = amt3.ToFinString()
	amt4, _ := amt1.Sub(&amt3)
	result["released_amount"] = amt4.ToFinString()

	// Whether to issue additional bitcoin for one-way transfer
	if lockbls_key[0] == 0 {
		trsno := big.NewInt(0).SetBytes(lockbls_key).Uint64()
		txhx, _ := state.ReadMoveBTCTxHashByTrsNo(uint32(trsno))
		if len(txhx) == 32 {
			result["satoshi_genesis_tx_hash"] = hex.EncodeToString(txhx)
		}
	}

	return result
}
