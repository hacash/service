package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/stores"
	"math/big"
	"strconv"
)

// 查询线性锁仓信息
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

	// 查询
	state := api.blockchain.State()
	lockblsItem := state.Lockbls(lockbls_key)
	if lockblsItem == nil {
		result["err"] = "not find."
		return result
	}

	// 返回信息
	result["master_address"] = lockblsItem.MasterAddress.ToReadable()
	result["effect_block_height"] = strconv.FormatUint(uint64(lockblsItem.EffectBlockHeight), 10)
	result["linear_block_number"] = strconv.FormatUint(uint64(lockblsItem.LinearBlockNumber), 10)
	amt1, _ := lockblsItem.GetAmount(&lockblsItem.TotalLockAmountBytes)
	result["total_lock_amount"] = amt1.ToFinString()
	amt2, _ := lockblsItem.GetAmount(&lockblsItem.LinearReleaseAmountBytes)
	result["linear_release_amount"] = amt2.ToFinString()
	amt3, _ := lockblsItem.GetAmount(&lockblsItem.BalanceAmountBytes)
	result["balance_amount"] = amt3.ToFinString()
	amt4, _ := amt1.Sub(amt3)
	result["released_amount"] = amt4.ToFinString()

	// 是否为单向转移比特币增发
	if lockbls_key[0] == 0 {
		trsno := big.NewInt(0).SetBytes(lockbls_key).Uint64()
		txhx, _ := state.ReadMoveBTCTxHashByNumber(uint32(trsno))
		if len(txhx) == 32 {
			result["satoshi_genesis_tx_hash"] = hex.EncodeToString(txhx)
		}
	}

	return result
}
