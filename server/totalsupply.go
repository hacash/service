package rpc

import (
	"fmt"
	"github.com/hacash/core/stores"
	"strconv"
	"strings"
)

// 输出 total supply 接口对象
func RenderTotalSupplyObject(totalsupply *stores.TotalSupply, isformatstring bool) (map[string]interface{}, map[string]string) {
	var ifs = isformatstring
	var object = make(map[string]interface{})
	var objstr = make(map[string]string)

	// 状态统计
	appendToUint64(object, objstr, "minted_diamond", uint64(totalsupply.Get(stores.TotalSupplyStoreTypeOfDiamond)), ifs)
	appendToUint64(object, objstr, "transferred_bitcoin", uint64(totalsupply.Get(stores.TotalSupplyStoreTypeOfTransferBitcoin)), ifs)
	// 钻石系统借贷，实时抵押中的钻石数量
	appendToUint64(object, objstr, "syslend_diamond_in_pledge", uint64(totalsupply.Get(stores.TotalSupplyStoreTypeOfSystemLendingDiamondCurrentMortgageCount)), ifs)
	// 比特币系统借贷，实时抵押中的比特币数量（统计数据为份数，一份 = 0.01BTC），转换成比特币数量
	appendToFloat64(object, objstr, "syslend_bitcoin_in_pledge", totalsupply.Get(stores.TotalSupplyStoreTypeOfSystemLendingBitcoinPortionCurrentMortgageCount)/100, ifs)
	// 用户间借贷抵押钻石数量流水累计
	appendToUint64(object, objstr, "usrlend_mortgage_diamond_count", uint64(totalsupply.Get(stores.TotalSupplyStoreTypeOfUsersLendingCumulationDiamond)), ifs)
	// 用户间借贷抵押比特币数量流水累计
	appendToUint64(object, objstr, "usrlend_mortgage_bitcoin_count", uint64(totalsupply.Get(stores.TotalSupplyStoreTypeOfUsersLendingCumulationBitcoin)), ifs)
	// 用户间借贷借出HAC流水累计
	appendToFloat64(object, objstr, "usrlend_loan_hac_count", totalsupply.Get(stores.TotalSupplyStoreTypeOfUsersLendingCumulationHacAmount), ifs)

	// 统计 HAC 增发
	miner_reward,
		channel_interest,
		btcmove_subsidy,
		syslend_diamond_loan_hac_count,
		syslend_bitcoin_loan_hac_count :=
		// 挖矿 + 通道利息 + 比特币转移增发
		totalsupply.Get(stores.TotalSupplyStoreTypeOfBlockMinerReward),
		totalsupply.Get(stores.TotalSupplyStoreTypeOfChannelInterest),
		totalsupply.Get(stores.TotalSupplyStoreTypeOfBitcoinTransferUnlockSuccessed),
		// 统计借贷相关增发
		totalsupply.Get(stores.TotalSupplyStoreTypeOfSystemLendingDiamondCumulationLoanHacAmount), // 钻石系统借贷累计借出
		totalsupply.Get(stores.TotalSupplyStoreTypeOfSystemLendingBitcoinPortionCumulationLoanHacAmount) // 比特币系统借贷累计借出

	// 统计 HAC 减扣
	syslend_diamond_repay_hac_count,
		syslend_bitcoin_repay_hac_count :=
		totalsupply.Get(stores.TotalSupplyStoreTypeOfSystemLendingDiamondCumulationRansomHacAmount), // 钻石系统借贷累计赎回
		totalsupply.Get(stores.TotalSupplyStoreTypeOfSystemLendingBitcoinPortionCumulationRansomHacAmount) // 比特币系统借贷累计赎回

	// 统计 HAC 销毁
	burned_fee,
		syslend_bitcoin_burning_interest,
		usrlend_burning_interest :=
		totalsupply.Get(stores.TotalSupplyStoreTypeOfBurningFee),
		totalsupply.Get(stores.TotalSupplyStoreTypeOfSystemLendingBitcoinPortionBurningInterestHacAmount), // 比特币系统借贷销毁利息
		totalsupply.Get(stores.TotalSupplyStoreTypeOfUsersLendingBurningOnePercentInterestHacAmount) // 用户借贷销毁1%利息

	// 增发
	appendToFloat64(object, objstr, "miner_reward", miner_reward, ifs)
	appendToFloat64(object, objstr, "channel_interest", btcmove_subsidy, ifs)
	appendToFloat64(object, objstr, "btcmove_subsidy", btcmove_subsidy, ifs)
	appendToFloat64(object, objstr, "syslend_diamond_loan_hac_count", syslend_diamond_loan_hac_count, ifs)   // 钻石系统借贷借出HAC累计
	appendToFloat64(object, objstr, "syslend_diamond_repay_hac_count", syslend_diamond_repay_hac_count, ifs) // 钻石系统借贷赎回HAC累计
	appendToFloat64(object, objstr, "syslend_bitcoin_loan_hac_count", syslend_bitcoin_loan_hac_count, ifs)   // 比特币系统借贷借出HAC累计
	appendToFloat64(object, objstr, "syslend_bitcoin_repay_hac_count", syslend_bitcoin_repay_hac_count, ifs) // 比特币系统借贷赎回HAC累计

	// 销毁
	appendToFloat64(object, objstr, "burned_fee", burned_fee, ifs)
	appendToFloat64(object, objstr, "syslend_bitcoin_burning_interest", syslend_bitcoin_burning_interest, ifs)
	appendToFloat64(object, objstr, "usrlend_burning_interest", usrlend_burning_interest, ifs)

	// 计算实时流通量
	totalAddAmountNum := miner_reward + channel_interest + btcmove_subsidy                        // 总增发
	totalSubAmountNum := burned_fee + syslend_bitcoin_burning_interest + usrlend_burning_interest // 总销毁
	// 借贷相关实时流通量统计
	totalAddAmountNum += syslend_diamond_loan_hac_count + syslend_bitcoin_loan_hac_count
	totalSubAmountNum += syslend_diamond_repay_hac_count + syslend_bitcoin_repay_hac_count
	// 实时流通量
	current_circulation := totalAddAmountNum - totalSubAmountNum
	appendToFloat64(object, objstr, "current_circulation", current_circulation, ifs)

	// 统计
	// 位于通道链中的HAC
	located_in_channel := totalsupply.Get(stores.TotalSupplyStoreTypeOfLocatedInChannel)
	if located_in_channel < 0.00000001 {
		located_in_channel = 0
	}
	appendToFloat64(object, objstr, "located_in_channel", located_in_channel, ifs)

	// 返回
	return object, objstr
}

// 输出格式
func appendToUint64(object map[string]interface{}, objectString map[string]string, key string, num uint64, isformatstring bool) {
	if isformatstring {
		numstr := fmt.Sprintf("%d", num)
		objectString[key] = numstr
	} else {
		object[key] = num
	}
}

// 输出格式
func appendToFloat64(object map[string]interface{}, objectString map[string]string, key string, num float64, isformatstring bool) {
	if isformatstring {
		numstr := strings.TrimSuffix(fmt.Sprintf("%.4f", num), ".0000")
		objectString[key] = numstr
	} else {
		object[key] = float2float(num)
	}
}

func float2float(num float64) float64 {
	float_num, _ := strconv.ParseFloat(fmt.Sprintf("%.8f", num), 64)
	return float_num
}
