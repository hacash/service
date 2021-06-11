package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	rpc "github.com/hacash/service/server"
	"strconv"
	"strings"
)

const (
	tystrOpenClannel     = "channel open"
	tystrCloseClannel    = "channel close"
	tystrOpenUsrLending  = "user lending open"
	tystrCloseUsrLending = "user lending close"
)

// 扫描区块 获取所有通道开启交易
func (api *DeprecatedApiService) getAllOperateActionLogByBlockHeight(params map[string]string) map[string]string {
	result := make(map[string]string)
	block_height_str, ok1 := params["block_height"]
	if !ok1 {
		result["err"] = "param block_height must."
		return result
	}
	block_height, err2 := strconv.ParseUint(block_height_str, 10, 0)
	if err2 != nil {
		result["err"] = "param block_height format error."
		return result
	}

	state := api.blockchain.State()

	lastest, e3 := state.ReadLastestBlockHeadAndMeta()
	if e3 != nil {
		result["err"] = e3.Error()
		return result
	}

	// 判断区块高度
	if block_height <= 0 || block_height > lastest.GetHeight() {
		result["err"] = "block height not find."
		result["ret"] = "1" // 返回错误码
		return result
	}

	// 查询区块
	tarblock, e := rpc.LoadBlockWithCache(api.backend.BlockChain().State(), block_height)
	if e != nil {
		result["err"] = "read block data error."
		return result
	}

	// 开始扫描区块
	allOperateLogs := make([]string, 0, 4)
	transactions := tarblock.GetTransactions()
	for _, v := range transactions {
		if 0 == v.Type() { // coinbase
			continue
		}
		for _, act := range v.GetActions() {
			kid := act.Kind()
			// 通道相关
			if 2 == kid { // 开启通道
				act := act.(*actions.Action_2_OpenPaymentChannel)
				desstr := act.LeftAmount.ToFinString() +
					"," + act.RightAmount.ToFinString()
				appendOperateActionLog(&allOperateLogs,
					kid, tystrOpenClannel, act.ChannelId,
					act.LeftAddress, act.RightAddress,
					desstr)
			} else if 3 == kid { // 关闭通道
				act := act.(*actions.Action_3_ClosePaymentChannel)
				appendOperateActionLogEx(&allOperateLogs,
					kid, tystrCloseClannel, act.ChannelId,
					"-", "-", "")
			} else if 12 == kid { // 关闭通道
				act := act.(*actions.Action_12_ClosePaymentChannelBySetupAmount)
				desstr := act.LeftAmount.ToFinString() +
					"," + act.RightAmount.ToFinString()
				appendOperateActionLog(&allOperateLogs,
					kid, tystrCloseClannel, act.ChannelId,
					act.LeftAddress, act.RightAddress,
					desstr)
			}
			// 借贷相关
			if 19 == kid { // 用户间借贷
				act := act.(*actions.Action_19_UsersLendingCreate)
				desstrs := make([]string, 0, 2)
				if act.MortgageBitcoin.NotEmpty.Check() {
					desstrs = append(desstrs, fmt.Sprintf("%d SAT", act.MortgageBitcoin.ValueSAT))
				}
				if act.MortgageDiamondList.Count > 0 {
					desstrs = append(desstrs, fmt.Sprintf("%d HACD", act.MortgageDiamondList.Count))
				}
				desstrs = append(desstrs, fmt.Sprintf("loan: %s", act.LoanTotalAmount.ToFinString()))
				desstrs = append(desstrs, fmt.Sprintf("repay: %s", act.AgreedRedemptionAmount.ToFinString()))
				appendOperateActionLog(&allOperateLogs,
					kid, tystrOpenUsrLending, act.LendingID,
					act.MortgagorAddress, act.LendersAddress,
					"collateral: "+strings.Join(desstrs, ", "))
			} else if 20 == kid {
				act := act.(*actions.Action_20_UsersLendingRansom)
				desstr := fmt.Sprintf("redeem: %s", act.RansomAmount.ToFinString())
				if act.RansomAmount.IsEmpty() {
					desstr = "clear"
				}
				appendOperateActionLogEx(&allOperateLogs,
					kid, tystrCloseUsrLending, act.LendingID,
					"-", "-",
					desstr)
			}

		}
	}

	datasstr := strings.Join(allOperateLogs, "\",\"")
	if len(datasstr) > 0 {
		datasstr = "\"" + datasstr + "\""
	}

	// 返回
	result["jsondata"] = `{"timestamp":` + strconv.FormatUint(tarblock.GetTimestamp(), 10) + `,"datas":[` + datasstr + `]}`
	return result
}

// 添加日志条目
func appendOperateActionLog(logary *[]string, kid uint16, tystr string, dataid []byte, addr1 fields.Address, addr2 fields.Address, describe string) {
	appendOperateActionLogEx(logary, kid, tystr, dataid, addr1.ToReadable(), addr2.ToReadable(), describe)
}

func appendOperateActionLogEx(logary *[]string, kid uint16, tystr string, dataid []byte, addr1 string, addr2 string, describe string) {
	*logary = append(*logary, fmt.Sprintf(
		"%d|%s|%s|%s|%s|%s",
		kid, tystr, hex.EncodeToString(dataid),
		addr1, addr2,
		describe,
	))
}
