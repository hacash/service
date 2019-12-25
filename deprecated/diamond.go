package rpc

import (
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/stores"
	"strconv"
	"strings"
)

//////////////////////////////////////////////////////////////

func (api *DeprecatedApiService) showDiamondCreateTxs(params map[string]string) map[string]string {
	result := make(map[string]string)

	txs := api.txpool.GetDiamondCreateTxs()

	jsondata := []string{}
	for i, tx := range txs {
		for _, act := range tx.GetActions() {
			if dcact, ok := act.(*actions.Action_4_DiamondCreate); ok {
				fee := tx.GetFee()
				jsondata = append(jsondata, fmt.Sprintf(`"%d","%s","%s","%s","%s","%s"`, i+1, tx.Hash().ToHex(), tx.GetAddress().ToReadable(),
					dcact.Diamond, dcact.Address.ToReadable(), fee.ToFinString()))
				break
			}
		}
	}
	perhei := 0
	lastest, _ := api.backend.BlockChain().State().ReadLastestBlockHeadAndMeta()
	if lastest != nil {
		perhei = (int(lastest.GetHeight()) + 5) / 5 * 5
	}

	result["jsondata"] = `{"period":"` + strconv.Itoa(perhei) + `","datas":[[` + strings.Join(jsondata, "],[") + `]]}`
	return result
}

func (api *DeprecatedApiService) getDiamond(params map[string]string) map[string]string {
	result := make(map[string]string)
	dmstr, ok1 := params["name"]
	if !ok1 {
		result["err"] = "params name must."
		return result
	}

	state := api.blockchain.State()
	blockstore := state.BlockStore()

	var store *stores.DiamondSmelt = nil
	if dmnum, e := strconv.Atoi(dmstr); e == nil {
		store, _ = blockstore.ReadDiamondByNumber(uint32(dmnum))
	} else {
		if len(dmstr) != 6 {
			result["fail"] = "name format error."
			return result
		}
		store, _ = blockstore.ReadDiamond(fields.Bytes6(dmstr))
	}
	if store == nil {
		result["fail"] = "not find."
		return result
	}
	dmstr = string(store.Diamond)
	// get current belong
	sto2 := state.Diamond(fields.Bytes6(dmstr))
	if sto2 != nil {
		result["address"] = sto2.Address.ToReadable()
	} else {
		result["address"] = store.MinerAddress.ToReadable()
	}
	// ok
	result["block_height"] = strconv.FormatUint(uint64(store.ContainBlockHeight), 10)
	result["number"] = strconv.Itoa(int(store.Number))
	result["miner_address"] = store.MinerAddress.ToReadable()
	return result
}
