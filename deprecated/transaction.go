package rpc

import (
	"encoding/hex"
	"fmt"
	actions2 "github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/transactions"
	"strings"
)

// 通过 hx 获取交易简介
func (api *DeprecatedApiService) getTransactionIntro(params map[string]string) map[string]string {
	result := make(map[string]string)
	trsid, ok1 := params["id"]
	if !ok1 {
		result["err"] = "param id must."
		return result
	}
	var trshx []byte
	if txhx, e := hex.DecodeString(trsid); e == nil && len(txhx) == 32 {
		trshx = txhx
	} else {
		result["err"] = "transaction hash error."
		return result
	}
	// 查询交易

	store := api.blockchain.State().BlockStore()

	blkhei, trsresbytes, err := store.ReadTransactionBytesByHash(trshx)
	if err != nil {
		result["err"] = err.Error()
		return result
	}
	if trsresbytes == nil {
		result["err"] = "transaction not fond."
		return result
	}

	trsres, _, err := transactions.ParseTransaction(trsresbytes, 0)
	if err != nil {
		result["err"] = err.Error()
		return result
	}

	// 解析 actions
	var actions = trsres.GetActions()
	var actions_ary []string
	var actions_strings = ""
	for _, act := range actions {
		var kind = act.Kind()
		actstr := fmt.Sprintf(`{"k":%d`, kind)
		if kind == 1 {
			acc := act.(*actions2.Action_1_SimpleTransfer)
			actstr += fmt.Sprintf(`,"to":"%s","amount":"%s"`,
				acc.Address.ToReadable(),
				acc.Amount.ToFinString(),
			)
		} else if kind == 4 {
			acc := act.(*actions2.Action_4_DiamondCreate)
			actstr += fmt.Sprintf(`,"number":"%s","name":"%s","address":"%s"`,
				acc.Address.ToReadable(),
				acc.Diamond,
				acc.Address.ToReadable(),
			)
		}
		actstr += "}"
		actions_ary = append(actions_ary, actstr)
	}
	actions_strings = strings.Join(actions_ary, ",")
	// 交易返回数据
	txaddr := fields.Address(trsres.GetAddress())
	var txfee = trsres.GetFee()
	result["jsondata"] = fmt.Sprintf(
		`{"block":{"height":%d,"timestamp":%d},"type":%d,"address":"%s","fee":"%s","timestamp":%d,"actioncount":%d,"actions":[%s]`,
		blkhei,
		trsres.GetTimestamp(),
		trsres.Type(),
		txaddr.ToReadable(), // 主地址
		txfee.ToFinString(),
		trsres.GetTimestamp(),
		len(actions),
		actions_strings,
	)

	// 收尾并返回
	result["jsondata"] += "}"
	return result
}
