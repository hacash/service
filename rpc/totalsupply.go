package rpc

import (
	"fmt"
	"github.com/hacash/core/stores"
	"net/http"
)

func (api *RpcService) totalSupply(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	state := api.backend.BlockChain().State()

	// get
	ttspl, e1 := state.ReadTotalSupply()
	if e1 != nil {
		ResponseError(w, e1)
		return
	}

	//fmt.Println(ttspl.Serialize())

	// return
	data := ResponseCreateData("minted_diamond", int(ttspl.Get(stores.TotalSupplyStoreTypeOfDiamond)))

	data["transferred_bitcoin"] = fmt.Sprintf("%d", uint64(ttspl.Get(stores.TotalSupplyStoreTypeOfTransferBitcoin)))

	// 统计
	miner_reward,
		channel_interest,
		btcmove_subsidy,
		burned_fee :=
		ttspl.Get(stores.TotalSupplyStoreTypeOfBlockMinerReward),
		ttspl.Get(stores.TotalSupplyStoreTypeOfChannelInterest),
		ttspl.Get(stores.TotalSupplyStoreTypeOfBitcoinTransferUnlockSuccessed),
		ttspl.Get(stores.TotalSupplyStoreTypeOfBurningFee)

	data["miner_reward"] = miner_reward
	data["channel_interest"] = channel_interest
	data["btcmove_subsidy"] = btcmove_subsidy

	data["burned_fee"] = burned_fee

	// 计算
	data["current_circulation"] = miner_reward + channel_interest + btcmove_subsidy - burned_fee

	// 统计
	// 位于通道链中的HAC
	located_in_channel := ttspl.Get(stores.TotalSupplyStoreTypeOfLocatedInChannel)
	if located_in_channel < 0.00000001 {
		located_in_channel = 0
	}
	data["located_in_channel"] = located_in_channel

	// ok
	ResponseData(w, data)
}
