package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	"strconv"
	"strings"
)

//////////////////////////////////////////////////////////////

func (api *DeprecatedApiService) getChannel(params map[string]string) map[string]string {
	result := make(map[string]string)
	idstr, ok1 := params["ids"]
	if !ok1 {
		result["err"] = "params ids must."
		return result
	}

	idlist := strings.Split(idstr, ",")
	if len(idlist) == 0 {
		result["err"] = "params ids must."
		return result
	}

	total_amount := fields.NewEmptyAmount()
	for i := 0; i < len(idlist); i++ {
		idstr := idlist[i]

		if len(idstr) != 32 {
			result["fail"] = "id format error."
			return result
		}

		chanid, e0 := hex.DecodeString(idstr)
		if e0 != nil {
			result["fail"] = "id format error."
			return result
		}

		state := api.blockchain.GetChainEngineKernel().StateRead()
		store, _ := state.Channel(fields.ChannelId(chanid))

		if store == nil {
			result["fail"] = "not find."
			return result
		}

		totalamt, _ := store.LeftAmount.Add(&store.RightAmount)
		totalsat := store.LeftSatoshi.GetRealSatoshi() + store.RightSatoshi.GetRealSatoshi()

		if len(idlist) == 1 {
			// 只有一条数据则返回详情
			iabtrs := map[uint8]string{
				0: "normal",
				1: "left",
				2: "right",
			}
			result["status"] = strconv.Itoa(int(store.Status))
			result["belong_height"] = strconv.FormatUint(uint64(store.BelongHeight), 10)
			result["lock_block"] = strconv.FormatUint(uint64(store.ArbitrationLockBlock), 10)
			result["reuse_version"] = strconv.FormatUint(uint64(store.ReuseVersion), 10)
			result["interest_attribution"] = iabtrs[uint8(store.InterestAttribution)]
			result["left_address"] = store.LeftAddress.ToReadable()
			result["left_amount"] = store.LeftAmount.ToFinString()
			result["left_satoshi"] = strconv.FormatUint(uint64(store.LeftSatoshi.GetRealSatoshi()), 10)
			result["right_address"] = store.RightAddress.ToReadable()
			result["right_amount"] = store.RightAmount.ToFinString()
			result["right_satoshi"] = strconv.FormatUint(uint64(store.RightSatoshi.GetRealSatoshi()), 10)
			result["total_amount"] = totalamt.ToFinString()
			if store.Status == 1 {
				result["challenge_height"] = strconv.FormatUint(uint64(store.ChallengeLaunchHeight), 10)
				result["assert_bill_number"] = strconv.FormatUint(uint64(store.AssertBillAutoNumber), 10)
				result["assert_address"] = store.RightAddress.ToReadable()
				if store.AssertAddressIsLeftOrRight.Check() {
					result["assert_address"] = store.LeftAddress.ToReadable()
				}
				result["assert_amount"] = store.AssertAmount.ToFinString()
				result["assert_satoshi"] = strconv.FormatUint(uint64(store.AssertSatoshi.GetRealSatoshi()), 10)
			}
			if store.Status == 2 || store.Status == 3 {
				// 计算各自分配
				l1 := store.LeftFinalDistributionAmount
				r1, _ := totalamt.Sub(&l1)
				l2 := uint64(store.LeftFinalDistributionSatoshi.GetRealSatoshi())
				r2 := uint64(totalsat) - l2
				result["distribution"] = fmt.Sprintf("left: %sHAC %dSAT, right: %sHAC %dSAT",
					l1.ToFinString(), l2, r1.ToFinString(), r2,
				)
				result["left_final_distribution_amount"] = store.LeftFinalDistributionAmount.ToFinString()
				result["left_final_distribution_satoshi"] = strconv.FormatUint(uint64(store.LeftFinalDistributionSatoshi.GetRealSatoshi()), 10)
			}
			return result
		} else {
			// 否则返回加总统计
			total_amount, _ = total_amount.Add(totalamt)
		}
	}

	// 返回总计
	result["total"] = strconv.Itoa(len(idlist))
	result["total_amount"] = total_amount.ToFinString()

	return result
}
