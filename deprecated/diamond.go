package rpc

import (
	"github.com/hacash/core/fields"
	"strconv"
)

//////////////////////////////////////////////////////////////

func (api *DeprecatedApiService) getDiamond(params map[string]string) map[string]string {
	result := make(map[string]string)
	dmstr, ok1 := params["name"]
	if !ok1 {
		result["err"] = "params name must."
		return result
	}
	if len(dmstr) != 6 {
		result["fail"] = "name format error."
		return result
	}

	state := api.blockchain.State()
	blockstore := state.BlockStore()

	store, e1 := blockstore.ReadDiamond(fields.Bytes6(dmstr))
	if e1 != nil || store == nil {
		result["fail"] = "not find."
		return result
	}

	sto2 := state.Diamond(fields.Bytes6(dmstr))
	if sto2 == nil {
		result["fail"] = "not find."
		return result
	}

	// 0
	result["block_height"] = strconv.FormatUint(uint64(store.ContainBlockHeight), 10)
	result["number"] = strconv.Itoa(int(store.Number))
	result["miner_address"] = store.MinerAddress.ToReadable()
	result["address"] = sto2.Address.ToReadable()
	return result
}
