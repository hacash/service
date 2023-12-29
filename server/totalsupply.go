package rpc

import (
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/stores"
	"strconv"
	"strings"
)

// Output total supply interface object
func RenderTotalSupplyObject(state interfaces.ChainStateOperationRead, totalsupply *stores.TotalSupply, isformatstring bool) (map[string]interface{}, map[string]string) {
	var ifs = isformatstring
	var object = make(map[string]interface{})
	var objstr = make(map[string]string)

	// Status statistics
	appendToUint64(object, objstr, "lastest_block_height", state.GetPendingBlockHeight()-1, ifs)
	appendToUint64(object, objstr, "minted_diamond", totalsupply.GetUint(stores.TotalSupplyStoreTypeOfDiamond), ifs)
	appendToUint64(object, objstr, "transferred_bitcoin", totalsupply.GetUint(stores.TotalSupplyStoreTypeOfTransferBitcoin), ifs)
	// Diamond system loan, real-time mortgage of diamond quantity
	appendToUint64(object, objstr, "syslend_diamond_in_pledge", totalsupply.GetUint(stores.TotalSupplyStoreTypeOfSystemLendingDiamondCurrentMortgageCount), ifs)
	// Bitcoin system lending, the number of bitcoins in real-time mortgage (the statistical data is the number of copies, one copy = 0.01btc), converted into the number of bitcoins
	appendToFloat64(object, objstr, "syslend_bitcoin_in_pledge", float64(totalsupply.GetUint(stores.TotalSupplyStoreTypeOfSystemLendingBitcoinPortionCurrentMortgageCount))/100.0, ifs)
	// Daily accumulation of inter user loan mortgage diamond quantity
	appendToUint64(object, objstr, "usrlend_mortgage_diamond_count", totalsupply.GetUint(stores.TotalSupplyStoreTypeOfUsersLendingCumulationDiamond), ifs)
	// Daily accumulation of inter user loan mortgage bitcoin quantity
	appendToUint64(object, objstr, "usrlend_mortgage_bitcoin_count", totalsupply.GetUint(stores.TotalSupplyStoreTypeOfUsersLendingCumulationBitcoin), ifs)
	// HAC flow accumulation of inter user loan and borrowing
	appendToFloat64(object, objstr, "usrlend_loan_hac_count", totalsupply.Get(stores.TotalSupplyStoreTypeOfUsersLendingCumulationHacAmount), ifs)

	// Statistical HAC additional issuance
	block_reward,
		channel_interest,
		btcmove_subsidy,
		syslend_diamond_loan_hac_count,
		syslend_bitcoin_loan_hac_count :=
		// Mining + channel interest + bitcoin transfer and issuance
		totalsupply.Get(stores.TotalSupplyStoreTypeOfBlockReward),
		totalsupply.Get(stores.TotalSupplyStoreTypeOfChannelInterest),
		totalsupply.Get(stores.TotalSupplyStoreTypeOfBitcoinTransferUnlockSuccessed),
		// Statistics of additional issuance related to lending
		totalsupply.Get(stores.TotalSupplyStoreTypeOfSystemLendingDiamondCumulationLoanHacAmount), // Cumulative lending of diamond system loans
		totalsupply.Get(stores.TotalSupplyStoreTypeOfSystemLendingBitcoinPortionCumulationLoanHacAmount) // Cumulative lending of debit and credit in bitcoin system

	// Statistical HAC deduction
	syslend_diamond_repay_hac_count,
		syslend_bitcoin_repay_hac_count :=
		totalsupply.Get(stores.TotalSupplyStoreTypeOfSystemLendingDiamondCumulationRansomHacAmount), // Cumulative redemption of diamond system loans
		totalsupply.Get(stores.TotalSupplyStoreTypeOfSystemLendingBitcoinPortionCumulationRansomHacAmount) // Cumulative redemption of debit and credit in bitcoin system

	// Statistical HAC destruction
	burned_fee,
		burned_hacd_bid,
		syslend_bitcoin_burning_interest,
		usrlend_burning_interest :=
		totalsupply.Get(stores.TotalSupplyStoreTypeOfBurningFeeTotal),
		totalsupply.GetUint(stores.TotalSupplyStoreTypeOfDiamondBidBurningZhu)/100000000,
		totalsupply.Get(stores.TotalSupplyStoreTypeOfSystemLendingBitcoinPortionBurningInterestHacAmount), // Bitcoin system loan destruction interest
		totalsupply.Get(stores.TotalSupplyStoreTypeOfUsersLendingBurningOnePercentInterestHacAmount) // User loan destroyed 1% interest

	// Additional issue
	appendToFloat64(object, objstr, "block_reward", block_reward, ifs)
	appendToFloat64(object, objstr, "channel_interest", channel_interest, ifs)
	appendToFloat64(object, objstr, "btcmove_subsidy", btcmove_subsidy, ifs)
	appendToFloat64(object, objstr, "syslend_diamond_loan_hac_count", syslend_diamond_loan_hac_count, ifs)   // HAC accumulation of diamond system lending and borrowing
	appendToFloat64(object, objstr, "syslend_diamond_repay_hac_count", syslend_diamond_repay_hac_count, ifs) // HAC accumulation of diamond system loan redemption
	appendToFloat64(object, objstr, "syslend_bitcoin_loan_hac_count", syslend_bitcoin_loan_hac_count, ifs)   // HAC accumulation of debit and credit in bitcoin system
	appendToFloat64(object, objstr, "syslend_bitcoin_repay_hac_count", syslend_bitcoin_repay_hac_count, ifs) // HAC accumulation of debit and credit redemption in bitcoin system

	// Destruction
	appendToFloat64(object, objstr, "burned_fee", burned_fee, ifs)
	appendToUint64(object, objstr, "burned_hacd_bid", burned_hacd_bid, ifs)
	appendToFloat64(object, objstr, "syslend_bitcoin_burning_interest", syslend_bitcoin_burning_interest, ifs)
	appendToFloat64(object, objstr, "usrlend_burning_interest", usrlend_burning_interest, ifs)

	// Calculate real-time circulation
	totalAddAmountNum := block_reward + channel_interest + btcmove_subsidy                        // 总增发
	totalSubAmountNum := burned_fee + syslend_bitcoin_burning_interest + usrlend_burning_interest // 总销毁

	// Loan related real-time Circulation Statistics
	totalAddAmountNum += syslend_diamond_loan_hac_count + syslend_bitcoin_loan_hac_count
	totalSubAmountNum += syslend_diamond_repay_hac_count + syslend_bitcoin_repay_hac_count

	// Real time circulation
	current_circulation := totalAddAmountNum - totalSubAmountNum
	appendToFloat64(object, objstr, "current_circulation", current_circulation, ifs)

	// Statistics
	// HAC in channel chain
	located_in_channel := totalsupply.Get(stores.TotalSupplyStoreTypeOfLocatedHACInChannel)
	channel_of_opening := totalsupply.GetUint(stores.TotalSupplyStoreTypeOfChannelOfOpening)
	if located_in_channel < 0.00000001 {
		located_in_channel = 0
	}
	appendToFloat64(object, objstr, "located_in_channel", located_in_channel, ifs)
	appendToUint64(object, objstr, "channel_of_opening", uint64(channel_of_opening), ifs)

	// return
	return object, objstr
}

// Output format
func appendToUint64(object map[string]interface{}, objectString map[string]string, key string, num uint64, isformatstring bool) {
	if isformatstring {
		numstr := fmt.Sprintf("%d", num)
		objectString[key] = numstr
	} else {
		object[key] = num
	}
}

// Output format
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
