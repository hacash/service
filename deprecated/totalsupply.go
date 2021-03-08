package rpc

import (
	"fmt"
	"github.com/hacash/core/stores"
	"strings"
)

func (api *DeprecatedApiService) totalSupply(params map[string]string) map[string]string {
	result := make(map[string]string)
	state := api.blockchain.State()

	// get
	ttspl, e1 := state.ReadTotalSupply()
	if e1 != nil {
		result["err"] = e1.Error()
		return result
	}

	//fmt.Println(ttspl.Serialize())

	// return
	result["minted_diamond"] = fmt.Sprintf("%d", uint64(ttspl.Get(stores.TotalSupplyStoreTypeOfDiamond)))
	result["transferred_bitcoin"] = fmt.Sprintf("%d", uint64(ttspl.Get(stores.TotalSupplyStoreTypeOfTransferBitcoin)))

	// 统计
	miner_reward,
		channel_interest,
		btcmove_subsidy,
		burned_fee :=
		ttspl.Get(stores.TotalSupplyStoreTypeOfBlockMinerReward),
		ttspl.Get(stores.TotalSupplyStoreTypeOfChannelInterest),
		ttspl.Get(stores.TotalSupplyStoreTypeOfBitcoinTransferUnlockSuccessed),
		ttspl.Get(stores.TotalSupplyStoreTypeOfBurningFee)

	result["miner_reward"] = strings.TrimSuffix(fmt.Sprintf("%.4f", miner_reward), ".0000")
	result["channel_interest"] = strings.TrimSuffix(fmt.Sprintf("%.4f", channel_interest), ".0000")
	result["btcmove_subsidy"] = strings.TrimSuffix(fmt.Sprintf("%.4f", btcmove_subsidy), ".0000")

	result["burned_fee"] = strings.TrimSuffix(fmt.Sprintf("%.4f", burned_fee), ".0000")

	// 计算
	result["current_circulation"] = strings.TrimSuffix(fmt.Sprintf("%.4f", miner_reward+channel_interest+btcmove_subsidy-burned_fee), ".0000")

	// 统计
	// 位于通道链中的HAC
	located_in_channel := ttspl.Get(stores.TotalSupplyStoreTypeOfLocatedInChannel)
	if located_in_channel < 0.00000001 {
		located_in_channel = 0
	}
	result["located_in_channel"] = strings.TrimSuffix(fmt.Sprintf("%.4f", located_in_channel), ".0000")

	// ok
	return result

}
