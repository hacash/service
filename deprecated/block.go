package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfacev2"
	"github.com/hacash/core/transactions"
	"github.com/hacash/mint/coinbase"
	"strconv"
	"strings"
)

// 通过 高度 或 hx 获取区块简介
func (api *DeprecatedApiService) getBlockIntro(params map[string]string) map[string]string {
	result := make(map[string]string)
	var isgettxhxs = false // 是否获取区块交易hash列表
	if _, ok0 := params["gettrshxs"]; ok0 {
		isgettxhxs = true
	}
	blkid, ok1 := params["id"]
	if !ok1 {
		result["err"] = "param id must."
		return result
	}

	store := api.blockchain.StateRead().BlockStoreRead()
	var err error

	var blockhx = []byte{}
	var blockbytes = []byte{}
	if blkhei, err := strconv.ParseUint(blkid, 10, 0); err == nil {
		blockhx, blockbytes, err = store.ReadBlockBytesByHeight(blkhei)
		if err != nil {
			result["err"] = err.Error()
			return result
		}
	} else if bhx, e := hex.DecodeString(blkid); e == nil && len(bhx) == fields.HashSize {
		blockhx = bhx
		blockbytes, err = store.ReadBlockBytesByHash(bhx)
		if err != nil {
			result["err"] = err.Error()
			return result
		}
	} else {
		result["err"] = "block id <" + blkid + "> not find."
		result["ret"] = "1"
		return result
	}
	// 解析区块
	var tarblock interfacev2.Block
	if isgettxhxs {
		tarblock, _, err = blocks.ParseBlock(blockbytes, 0)
	} else {
		tarblock, _, err = blocks.ParseBlockHead(blockbytes, 0)
	}
	if err != nil {
		result["err"] = err.Error()
		return result
	}
	// 区块返回数据
	result["jsondata"] = fmt.Sprintf(
		`{"hash":"%s","height":%d,"prevhash":"%s","mrklroot":"%s","timestamp":%d,"txcount":%d,"reward":"%s"`,
		hex.EncodeToString(blockhx),
		tarblock.GetHeight(),
		hex.EncodeToString(tarblock.GetPrevHash()),
		hex.EncodeToString(tarblock.GetMrklRoot()),
		tarblock.GetTimestamp(),
		tarblock.GetTransactionCount(),
		coinbase.BlockCoinBaseReward(tarblock.GetHeight()).ToFinString(), // 奖励数量
	)
	// 区块hx列表
	if isgettxhxs {
		var blktxhxsary []string
		var blktxhxsstr = ""
		var rwdaddr fields.Address // 奖励地址
		var rwdmsg string
		for i, trs := range tarblock.GetTransactions() {
			if i == 0 {
				rwdmsg = trs.GetMessage().ValueShow()
				rwdaddr = trs.GetAddress()
				blktxhxsary = append(blktxhxsary, "[coinbase]")
			} else {
				blktxhxsary = append(blktxhxsary, hex.EncodeToString(trs.Hash()))
			}
		}
		blktxhxsstr = strings.Join(blktxhxsary, `","`)
		if len(blktxhxsstr) > 0 {
			blktxhxsstr = `"` + blktxhxsstr + `"`
		}
		result["jsondata"] += fmt.Sprintf(
			`,"nonce":%d,"difficulty":%d,"rwdaddr":"%s","rwdmsg":"%s","trshxs":[%s]`,
			tarblock.GetNonce(),
			tarblock.GetDifficulty(),
			rwdaddr.ToReadable(),
			rwdmsg,
			blktxhxsstr,
		)
	}
	// 收尾并返回
	result["jsondata"] += "}"
	return result
}

// 获取最新区块高度
func (api *DeprecatedApiService) getLastBlockHeight(params map[string]string) map[string]string {
	result := make(map[string]string)

	state := api.blockchain.StateRead()
	lastest, err := state.ReadLastestBlockHeadMetaForRead()
	if err != nil {
		result["err"] = err.Error()
		return result
	}

	result["jsondata"] = fmt.Sprintf(
		`{"height":%d,"txs":%d,"timestamp":%d}`,
		lastest.GetHeight(),
		lastest.GetCustomerTransactionCount(),
		lastest.GetTimestamp(),
	)
	return result
}

// 获取区块摘要信息
func (api *DeprecatedApiService) getBlockAbstractList(params map[string]string) map[string]string {
	result := make(map[string]string)
	start, ok1 := params["start_height"]
	end, ok2 := params["end_height"]
	if !ok1 || !ok2 {
		result["err"] = "start_height or end_height must"
		return result
	}
	start_hei, e1 := strconv.ParseUint(start, 10, 0)
	end_hei, e2 := strconv.ParseUint(end, 10, 0)
	if e1 != nil || e2 != nil {
		result["err"] = "start_height or end_height param error"
		return result
	}
	if end_hei-start_hei+1 > 100 {
		result["err"] = "start_height - end_height cannot more than 100"
		return result
	}
	// 查询区块信息

	store := api.blockchain.StateRead().BlockStoreRead()

	coinbase_start_pos := uint32(blocks.BlockHeadSize + blocks.BlockMetaSizeV1)
	coinbase_head_len := uint32(1 + 21 + 3 + 16 + 1)
	var jsondata = make([]string, 0, end_hei-start_hei+1)
	for i := end_hei; i >= start_hei; i-- {
		//blkhash, blkbytes, e := store.ReadBlockBytesLengthByHeight(i, coinbase_start_pos+coinbase_head_len)
		blkhash, blkbytes, e := store.ReadBlockBytesByHeight(i)
		if e != nil {
			result["err"] = e.Error()
			return result
		}
		if blkhash == nil || blkbytes == nil {
			result["err"] = "block height not find. " + fmt.Sprintf("", coinbase_head_len)
			return result
		}
		blkhead, _, e2 := blocks.ParseExcludeTransactions(blkbytes, 0)
		if e2 != nil {
			result["err"] = e2.Error()
			return result
		}
		// 解析矿工信息
		cbtx, _, e := transactions.ParseTransaction(blkbytes, coinbase_start_pos)
		if e != nil {
			result["err"] = e.Error()
			return result
		}
		// 返回
		msg := cbtx.GetMessage().ValueShow()
		jsondata = append(jsondata, fmt.Sprintf(
			`{"hash":"%s","txs":%d,"time":%d,"height":%d,"nonce":%d,"bits":%d,"rewards":{"amount":"%s","address":"%s","message":"%s"}}`,
			hex.EncodeToString(blkhash),
			blkhead.GetCustomerTransactionCount(),
			blkhead.GetTimestamp(),
			blkhead.GetHeight(),
			blkhead.GetNonce(),
			blkhead.GetDifficulty(),
			coinbase.BlockCoinBaseReward(blkhead.GetHeight()).ToFinString(),
			cbtx.GetAddress().ToReadable(),
			msg,
		))
		//addrbytes = bytes.Trim(addrbytes, string([]byte{0}))
		//fmt.Println([]byte(coinbase.Message))
		//fmt.Println(i, cbtx.GetAddress().ToReadable())
	}
	// 返回
	result["jsondata"] = `{"datas":[` + strings.Join(jsondata, ",") + `]}`
	return result
}
