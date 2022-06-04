package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/stores"
	"strconv"
)

//////////////////////////////////////////////////////////////
func (api *DeprecatedApiService) getUserLending(params map[string]string) map[string]string {
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

	if len(dataid) != stores.UserLendingIdLength {
		result["fail"] = "id length error."
		return result
	}
	state := api.blockchain.GetChainEngineKernel().StateRead()

	stoobj, _ := state.UserLending(fields.UserLendingId(dataid))
	if stoobj == nil {
		result["fail"] = "not find."
		return result
	}

	// Return details
	result["is_ransomed"] = strconv.Itoa(int(stoobj.IsRansomed.Value()))
	result["is_redemption_overtime"] = strconv.Itoa(int(stoobj.IsRedemptionOvertime.Value()))
	result["is_public_redeemable"] = strconv.Itoa(int(stoobj.IsPublicRedeemable.Value()))
	result["create_block_height"] = strconv.FormatUint(uint64(stoobj.CreateBlockHeight), 10)
	result["expire_block_height"] = strconv.FormatUint(uint64(stoobj.ExpireBlockHeight), 10)
	result["mortgagor_address"] = stoobj.MortgagorAddress.ToReadable()
	result["lender_address"] = stoobj.LenderAddress.ToReadable()

	result["mortgage_satoshi"] = strconv.FormatUint(uint64(stoobj.MortgageBitcoin.ValueSAT), 10)
	result["mortgage_diamonds"] = stoobj.MortgageDiamondList.SerializeHACDlistToCommaSplitString() // Diamond list
	result["loan_amount"] = stoobj.LoanTotalAmount.ToFinString()
	result["repay_amount"] = stoobj.AgreedRedemptionAmount.ToFinString()
	result["burned_interest_amount"] = stoobj.PreBurningInterestAmount.ToFinString()

	// Show return status
	if stoobj.IsRansomed.Check() {
		result["ransom_block_height"] = strconv.FormatUint(uint64(stoobj.RansomBlockHeight), 10)
		result["ransom_amount"] = stoobj.RansomAmount.ToFinString()
		result["ransom_address"] = stoobj.RansomAddress.ToReadable() // display address
	}

	// return
	return result
}
