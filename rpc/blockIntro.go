package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/transactions"
	"github.com/hacash/mint/coinbase"
	"net/http"
)

func (api *RpcService) blockIntro(r *http.Request, w http.ResponseWriter) {

	height, ok0 := CheckParamUint64Must(r, w, "height")
	if !ok0 {
		return
	}

	// 是否以枚为单位
	isUnitMei := CheckParamBool(r, "unitmei", false)

	// 区块储存
	blkstore := api.backend.BlockChain().State().BlockStore()

	// get
	coinbase_start_pos := uint32(blocks.BlockHeadSize + blocks.BlockMetaSizeV1)
	readlen := coinbase_start_pos + uint32(1+21+3+16+1) // coinbase len
	blkhash, blkbytes, e1 := blkstore.ReadBlockBytesByHeight(height, readlen)
	if e1 != nil {
		ResponseError(w, e1)
		return
	}
	block, _, e2 := blocks.ParseExcludeTransactions(blkbytes, 0)
	if e2 != nil {
		ResponseError(w, e2)
		return
	}

	// 解析矿工信息
	cbtx, _, e3 := transactions.ParseTransaction(blkbytes, coinbase_start_pos)
	if e3 != nil {
		ResponseError(w, e3)
		return
	}

	blockHeight := block.GetHeight()

	// head
	data := ResponseCreateData("version", block.Version())
	data["height"] = blockHeight
	data["timestamp"] = block.GetTimestamp()
	data["hash"] = hex.EncodeToString(blkhash)
	data["prev_hash"] = block.GetPrevHash().ToHex()
	data["mrkl_root"] = block.GetMrklRoot().ToHex()
	data["transaction_count"] = block.GetCustomerTransactionCount() // drop coinbase trs

	// meta
	data["nonce"] = block.GetNonce()
	data["difficulty"] = block.GetDifficulty()

	// coinbase
	var coinbaseitem = map[string]interface{}{}
	coinbaseitem["address"] = cbtx.GetAddress().ToReadable()
	rewardAmt := coinbase.BlockCoinBaseReward(blockHeight)
	if isUnitMei {
		coinbaseitem["reward"] = rewardAmt.ToMeiString(16)
	} else {
		coinbaseitem["reward"] = rewardAmt.ToFinString()
	}
	data["coinbase"] = coinbaseitem

	// return
	ResponseData(w, data)
}
