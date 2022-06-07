package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/transactions"
	"github.com/hacash/mint/coinbase"
	"net/http"
)

func (api *RpcService) blockIntro(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	var errread error
	var blkhash []byte = nil
	var blkbytes []byte = nil

	height := CheckParamUint64(r, "height", 0)
	hashStr := CheckParamString(r, "hash", "")
	if len(hashStr) > 0 {
		blkhash, errread = hex.DecodeString(hashStr)
		if errread != nil {
			ResponseError(w, errread)
			return
		}
		if len(blkhash) != 32 {
			ResponseErrorString(w, "param <hash> format error")
			return
		}
	}

	// Is it in pieces
	isUnitMei := CheckParamBool(r, "unitmei", false)

	// Block Storage 
	blkstore := api.backend.BlockChain().GetChainEngineKernel().StateRead().BlockStoreRead()

	// get coinbase
	coinbase_start_pos := uint32(blocks.BlockHeadSize + blocks.BlockMetaSizeV1)
	//readlen := coinbase_start_pos + uint32(1+21+3+16+1) // coinbase len

	// Number of reads
	if len(blkhash) == 32 {
		blkbytes, errread = blkstore.ReadBlockBytesByHash(blkhash)
	} else {
		blkhash, blkbytes, errread = blkstore.ReadBlockBytesByHeight(height)
	}

	// Check for errors
	if errread != nil {
		ResponseError(w, errread)
		return
	}

	if blkbytes == nil {
		ResponseError(w, fmt.Errorf("block is not find"))
		return
	}

	// Parsing block information
	block, _, e2 := blocks.ParseExcludeTransactions(blkbytes, 0)
	if e2 != nil {
		ResponseError(w, e2)
		return
	}

	// Analyze miner information
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
	coinbaseitem["reward"] = rewardAmt.ToMeiOrFinString(isUnitMei)
	data["coinbase"] = coinbaseitem

	// return
	ResponseData(w, data)
}
