package rpc

import (
	"encoding/hex"
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
	if len(lockbls_key) != 24 {
		result["err"] = "param lockbls_id length error."
		return result
	}

	// 查询
	lockblsItem := api.blockchain.State().Lockbls(lockbls_key)
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
	return result
}
