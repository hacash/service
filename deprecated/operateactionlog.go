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
	tystrOpenClannel         = "channel open"
	tystrCloseClannel        = "channel close"
	tystrOpenUsrLending      = "user lending open"
	tystrCloseUsrLending     = "user lending close"
	tystrOpenDiamondLending  = "diamond syslend open"
	tystrCloseDiamondLending = "diamond syslend close"
	tystrOpenBitcoinLending  = "bitcoin syslend open"
	tystrCloseBitcoinLending = "bitcoin syslend close"

	tystrBitcoinMove = "bitcoin move"
	tystrOpenLockbls = "lockbls open"
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

	lastest, _, e3 := api.blockchain.GetChainEngineKernel().LatestBlock()
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
	tarblock, e := rpc.LoadBlockWithCache(api.backend.BlockChain().GetChainEngineKernel(), block_height)
	if e != nil {
		result["err"] = "read block data error."
		return result
	}

	// 开始扫描区块
	allOperateLogs := make([]string, 0, 4)
	transactions := tarblock.GetTrsList()
	for _, v := range transactions {
		if 0 == v.Type() { // coinbase
			continue
		}

		mainAddressString := v.GetAddress().ToReadable()
		for _, act := range v.GetActionList() {
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
					mainAddressString, "-", "")

			} else if 12 == kid { // 关闭通道
				act := act.(*actions.Action_12_ClosePaymentChannelBySetupAmount)
				desstr := act.LeftAmount.ToFinString() +
					"," + act.RightAmount.ToFinString()
				appendOperateActionLog(&allOperateLogs,
					kid, tystrCloseClannel, act.ChannelId,
					act.LeftAddress, act.RightAddress,
					desstr)
			}
			// 比特币转移和锁仓
			if 7 == kid { // 比特币转移
				act := act.(*actions.Action_7_SatoshiGenesis)
				dataID := actions.GainLockblsIdByBtcMove(uint32(act.TransferNo))
				desstr := fmt.Sprintf("move: %d BTC, reward: ㄜ%d:248",
					act.BitcoinQuantity,
					act.AdditionalTotalHacAmount)
				appendOperateActionLogEx(&allOperateLogs,
					kid, tystrBitcoinMove, dataID,
					act.OriginAddress.ToReadable(), "-",
					desstr)

			} else if 9 == kid { // 线性锁仓，创建
				act := act.(*actions.Action_9_LockblsCreate)
				desstr := fmt.Sprintf("lock: %s, release: %s, step: %d",
					act.TotalStockAmount.ToFinString(),
					act.LinearReleaseAmount.ToFinString(),
					act.LinearBlockNumber)
				appendOperateActionLog(&allOperateLogs,
					kid, tystrOpenLockbls, act.LockblsId,
					act.PaymentAddress, act.MasterAddress,
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
					act.MortgagorAddress, act.LenderAddress,
					"collateral: "+strings.Join(desstrs, ", "))

			} else if 20 == kid { // 赎回或清算用户间借贷
				act := act.(*actions.Action_20_UsersLendingRansom)
				desstr := fmt.Sprintf("redeem: %s", act.RansomAmount.ToFinString())
				if act.RansomAmount.IsEmpty() {
					desstr = "clear"
				}
				appendOperateActionLogEx(&allOperateLogs,
					kid, tystrCloseUsrLending, act.LendingID,
					mainAddressString, "-",
					desstr)

			} else if 15 == kid { // 钻石系统借贷 开启
				act := act.(*actions.Action_15_DiamondsSystemLendingCreate)
				desstr := fmt.Sprintf("mortgage: %d HACD, loan: %s, interest: %.1f%%",
					act.MortgageDiamondList.Count,
					act.LoanTotalAmount.ToFinString(),
					float32(act.BorrowPeriod)*0.5)
				appendOperateActionLogEx(&allOperateLogs,
					kid, tystrOpenDiamondLending, act.LendingID,
					mainAddressString, "-",
					desstr)

			} else if 16 == kid { // 钻石系统借贷 赎回
				act := act.(*actions.Action_16_DiamondsSystemLendingRansom)
				desstr := fmt.Sprintf("redeem: %s",
					act.RansomAmount.ToFinString())
				appendOperateActionLogEx(&allOperateLogs,
					kid, tystrCloseDiamondLending, act.LendingID,
					mainAddressString, "-",
					desstr)

			} else if 17 == kid { // 比特币系统借贷 开启
				act := act.(*actions.Action_17_BitcoinsSystemLendingCreate)
				desstr := fmt.Sprintf("mortgage:%.2f BTC, loan: %s, interest: %s",
					float64(act.MortgageBitcoinPortion)/100,
					act.LoanTotalAmount.ToFinString(),
					act.PreBurningInterestAmount.ToFinString())
				appendOperateActionLogEx(&allOperateLogs,
					kid, tystrOpenBitcoinLending, act.LendingID,
					mainAddressString, "-",
					desstr)

			} else if 18 == kid { // 比特币系统借贷 赎回
				act := act.(*actions.Action_18_BitcoinsSystemLendingRansom)
				desstr := fmt.Sprintf("redeem: %s",
					act.RansomAmount.ToFinString())
				appendOperateActionLogEx(&allOperateLogs,
					kid, tystrCloseBitcoinLending, act.LendingID,
					mainAddressString, "-",
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
