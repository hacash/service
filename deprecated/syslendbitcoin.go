package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/coinbase"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/stores"
	"strconv"
)

//////////////////////////////////////////////////////////////

func (api *DeprecatedApiService) getSystemLendingBitcoin(params map[string]string) map[string]string {
	result := make(map[string]string)
	idstr, ok1 := params["id"]
	if !ok1 {
		result["err"] = "params id must."
		return result
	}

	dataid, e0 := hex.DecodeString(idstr)
	if e0 != nil {
		result["fail"] = "id format error."
		return result
	}

	if len(dataid) != stores.BitcoinSyslendIdLength {
		result["fail"] = "id length error."
		return result
	}
	state := api.blockchain.State()

	stoobj, _ := state.BitcoinSystemLending(fields.BitcoinSyslendId(dataid))
	if stoobj == nil {
		result["fail"] = "not find."
		return result
	}

	// 返回详情
	result["is_ransomed"] = strconv.Itoa(int(stoobj.IsRansomed.Value()))
	result["create_block_height"] = strconv.FormatUint(uint64(stoobj.CreateBlockHeight), 10)
	result["main_address"] = stoobj.MainAddress.ToReadable()
	result["mortgage_satoshi"] = strconv.FormatInt(int64(stoobj.MortgageBitcoinPortion)*100*10000, 10)
	result["loan_amount"] = stoobj.LoanTotalAmount.ToFinString()
	result["burned_interest_amount"] = stoobj.PreBurningInterestAmount.ToFinString()
	rtlper := float64(stoobj.RealtimeTotalMortgageRatio) / 100
	result["realtime_total_mortgage_ratio"] = fmt.Sprintf("%0.2f%%", rtlper) // 实时抵押比例
	loanbei, pbi := coinbase.CalculationOfInterestBitcoinMortgageLoanAmount(rtlper)
	result["realtime_interest_ratio"] = fmt.Sprintf("%0.2f%%", pbi/loanbei) // 实时利率
	result["realtime_loan_ratio"] = fmt.Sprintf("%0.2f%%", loanbei*100)     // 实时可借贷倍数
	// 显示归还状态
	if stoobj.IsRansomed.Check() {
		result["ransom_block_height"] = strconv.FormatUint(uint64(stoobj.RansomBlockHeight), 10)
		result["ransom_amount"] = stoobj.RansomAmount.ToFinString()
		result["ransom_address_if_public_operation"] = stoobj.RansomAddressIfPublicOperation.ShowReadableOrEmpty() // 如果存在则显示地址
	}
	// 返回
	return result
}
