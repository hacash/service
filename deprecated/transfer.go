package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/account"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/transactions"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// 创建一普通转账交易
func (api *DeprecatedApiService) transferSimple(params map[string]string) map[string]string {

	result := make(map[string]string)
	password_or_privatekey, ok1 := params["password"]
	if !ok1 {
		result["err"] = "password must"
		return result
	}
	var acc *account.Account = nil
	privatekey, e2 := hex.DecodeString(password_or_privatekey)
	if len(password_or_privatekey) == 64 && e2 == nil {
		acc, e2 = account.GetAccountByPriviteKey(privatekey)
		if e2 != nil {
			result["err"] = "Privite Key Error"
			return result
		}
	} else {
		//fmt.Println(password_or_privatekey)
		acc = account.CreateAccountByPassword(password_or_privatekey)
		//fmt.Println(string(acc.AddressReadable))
		//fmt.Println(params["from"])
	}
	if strings.Compare(string(acc.AddressReadable), params["from"]) != 0 {
		result["err"] = "Privite Key error with address " + params["from"]
		return result
	}
	// 私钥
	allPrivateKeyBytes := make(map[string][]byte, 1)
	allPrivateKeyBytes[string(acc.Address)] = acc.PrivateKey
	to_addr, e8 := fields.CheckReadableAddress(params["to"])
	if e8 != nil {
		result["err"] = e8.Error()
		return result
	}
	// 金额转换
	amount, e6 := fields.NewAmountFromFinString(params["amount"])
	if e6 != nil {
		result["err"] = e6.Error()
		return result
	}
	fee, e7 := fields.NewAmountFromFinString(params["fee"])
	if e7 != nil {
		result["err"] = e7.Error()
		return result
	}
	// 创建普通转账交易
	newTrs, e5 := transactions.NewEmptyTransaction_2_Simple(acc.Address)
	if e5 != nil {
		result["err"] = e5.Error()
		return result
	}
	curtime := time.Now()
	timestamp, ok := params["timestamp"]
	if ok {
		timebig, bb := new(big.Int).SetString(timestamp, 10)
		if !bb {
			result["err"] = "timestamp error"
			return result
		}
		curtime = time.Unix(timebig.Int64(), 0)
	}
	newTrs.Timestamp = fields.VarUint5(curtime.Unix()) // 使用当前时间戳
	newTrs.Fee = *fee                                  // set fee
	tranact := actions.NewAction_1_SimpleTransfer(*to_addr, amount)
	newTrs.AppendAction(tranact)
	// sign
	e9 := newTrs.FillNeedSigns(allPrivateKeyBytes, nil)
	if e9 != nil {
		result["err"] = e9.Error()
		return result
	}
	// 返回交易
	result["txhash"] = hex.EncodeToString(newTrs.HashFresh())
	txbody, e10 := newTrs.Serialize()
	if e10 != nil {
		result["err"] = e10.Error()
		return result
	}
	result["txbody"] = hex.EncodeToString(txbody)
	return result
}

// 查询交易确认状态
func (api *DeprecatedApiService) txStatus(params map[string]string) map[string]string {
	result := make(map[string]string)
	result["status"] = "" // 表示状态
	// 交易哈希
	txhashstr, ok1 := params["txhash"]
	if !ok1 {
		result["err"] = "txhash must"
		return result
	}
	txhash, e2 := hex.DecodeString(txhashstr)
	if len(txhashstr) != 64 || e2 != nil {
		result["err"] = "txhash format error"
		return result
	}

	state := api.blockchain.State()
	store := state.BlockStore()

	// 从交易池中读取
	_, ok2 := api.txpool.CheckTxExistByHash(txhash)
	if ok2 {
		// 交易正在交易池内
		result["status"] = "txpool" // 表示正在交易池
		return result
	}

	//miner := miner.GetGlobalInstanceHacashMiner()

	/* 从正在处理的区块中查询
	//fmt.Println(miner.CurrentPenddingBlock)
	if miner.CurrentPenddingBlock != nil {
		txs := miner.CurrentPenddingBlock.GetTransactions()
		for _, tx := range txs {
			if bytes.Compare(tx.HashNoFee(), txhash) == 0 {
				result["status"] = "block"                                                              // 表示正在挖矿的区块内
				result["block_height"] = strconv.FormatUint(miner.CurrentPenddingBlock.GetHeight(), 10) // 所属区块高度
				result["block_hash"] = hex.EncodeToString(miner.CurrentPenddingBlock.Hash())
				return result
			}
		}
	}
	*/

	//fmt.Println("GetGlobalInstanceBlocksDataStore")
	// 从区块数据中查询

	blkhei, txbody, e3 := store.ReadTransactionBytesByHash(txhash)
	if e3 != nil {
		result["err"] = e3.Error()
		return result
	}
	if txbody == nil {
		result["status"] = "notfind" // 表示不存在
		return result
	}
	// 存在并返回
	result["status"] = "confirm" // 表示不存在

	lastest, e4 := state.ReadLastestBlockHeadAndMeta()
	if e4 != nil {
		result["err"] = e4.Error()
		return result
	}

	// 查询区块hash
	tarblkhash, e5 := state.BlockStore().ReadBlockHashByHeight(blkhei)
	if e5 != nil {
		result["err"] = e5.Error()
		return result
	}

	confirm_height := lastest.GetHeight() - blkhei
	result["confirm_height"] = strconv.Itoa(int(confirm_height)) // 确认区块数
	result["block_height"] = strconv.FormatUint(blkhei, 10)      // 所属区块高度
	result["block_hash"] = tarblkhash.ToHex()                    // 所属区块hash
	return result
}
