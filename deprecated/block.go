package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"github.com/hacash/mint/coinbase"
	"strconv"
	"strings"
)

// Obtain block profile by height or HX
func (api *DeprecatedApiService) changeBlockReferHeight(params map[string]string) map[string]string {
	result := make(map[string]string)

	blkid, ok1 := params["id"]
	if !ok1 {
		result["err"] = "param id must."
		return result
	}
	store := api.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead()

	var tarblock interfaces.Block = nil
	if bhx, e := hex.DecodeString(blkid); e == nil && len(bhx) == fields.HashSize {
		blockbytes, err := store.ReadBlockBytesByHash(bhx)
		if err != nil {
			result["err"] = err.Error()
			return result
		}
		if len(blockbytes) > 0 {
			tarblock, _, err = blocks.ParseExcludeTransactions(blockbytes, 0)
			if err != nil {
				result["err"] = err.Error()
				return result
			}
		}
	}

	// if not find
	if tarblock == nil {
		result["err"] = "block hash <" + blkid + "> not find."
		result["ret"] = "1"
		return result
	}

	// to update
	state := api.blockchain.GetChainEngineKernel().CurrentState().BlockStore()
	err := state.UpdateSetBlockHashReferToHeight(tarblock.GetHeight(), tarblock.Hash())
	if err != nil {
		result["err"] = err.Error()
		return result
	}

	result["ok"] = fmt.Sprintf("Update block refer: %d => %s", tarblock.GetHeight(), tarblock.Hash().ToHex())

	return result

}

func (api *DeprecatedApiService) getBlockDataOfHex(params map[string]string) map[string]string {
	result := make(map[string]string)

	blkid, ok1 := params["id"]
	if !ok1 {
		result["err"] = "param id must."
		return result
	}

	isbody, _ := params["body"]

	store := api.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead()

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
	}
	// if not find
	if blockbytes == nil || len(blockbytes) == 0 {
		result["err"] = "block id <" + blkid + "> not find."
		result["ret"] = "1"
		return result
	}

	// Block return data
	if len(isbody) > 0 {
		result["jsondata"] = hex.EncodeToString(blockbytes)
	} else {
		result["jsondata"] = fmt.Sprintf(
			`{"hash":"%s","data":"%s"}`,
			hex.EncodeToString(blockhx),
			hex.EncodeToString(blockbytes),
		)
	}

	return result

}

// Obtain block profile by height or HX
func (api *DeprecatedApiService) getBlockIntro(params map[string]string) map[string]string {
	result := make(map[string]string)
	var isgettxhxs = false // Whether to obtain block transaction Hash list
	if _, ok0 := params["gettrshxs"]; ok0 {
		isgettxhxs = true
	}
	blkid, ok1 := params["id"]
	if !ok1 {
		result["err"] = "param id must."
		return result
	}

	state := api.blockchain.GetChainEngineKernel().StateRead()
	store := state.BlockStoreRead()
	var err error

	must_confirm, _ := params["must_confirm"]
	if len(must_confirm) > 0 {
		var okey_block_hei = state.GetPendingBlockHeight()
		//fmt.Println("okey_block_hei: ", okey_block_hei)
		must_confirm_block_hei, _ := strconv.ParseUint(must_confirm, 10, 0)
		get_blk_id, _ := strconv.ParseUint(blkid, 10, 0)
		if get_blk_id > 0 && must_confirm_block_hei > 0 && get_blk_id+must_confirm_block_hei > okey_block_hei {
			//fmt.Println("must_confirm_block_hei: ", must_confirm_block_hei)
			result["err"] = fmt.Sprintf("block %d not be confirm", get_blk_id)
			return result
		}
	}

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
	}
	// if not find
	if blockbytes == nil || len(blockbytes) == 0 {
		result["err"] = "block id <" + blkid + "> not find."
		result["ret"] = "1"
		return result
	}
	// Parsing block
	var tarblock interfaces.Block
	if isgettxhxs {
		tarblock, _, err = blocks.ParseBlock(blockbytes, 0)
	} else {
		tarblock, _, err = blocks.ParseBlockHead(blockbytes, 0)
	}
	if err != nil {
		result["err"] = err.Error()
		return result
	}
	// Block return data
	result["jsondata"] = fmt.Sprintf(
		`{"hash":"%s","height":%d,"prevhash":"%s","mrklroot":"%s","timestamp":%d,"txcount":%d,"reward":"%s"`,
		hex.EncodeToString(blockhx),
		tarblock.GetHeight(),
		hex.EncodeToString(tarblock.GetPrevHash()),
		hex.EncodeToString(tarblock.GetMrklRoot()),
		tarblock.GetTimestamp(),
		tarblock.GetTransactionCount(),
		coinbase.BlockCoinBaseReward(tarblock.GetHeight()).ToFinString(), // Number of awards
	)
	// Block HX list
	if isgettxhxs {
		var blktxhxsary []string
		var blktxhxsstr = ""
		var rwdaddr fields.Address // Reward address
		var rwdmsg string
		for i, trs := range tarblock.GetTrsList() {
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
	// Wrap up and return
	result["jsondata"] += "}"
	return result
}

// Get the latest block height
func (api *DeprecatedApiService) getLastBlockHeight(params map[string]string) map[string]string {
	result := make(map[string]string)

	lastest, _, err := api.blockchain.GetChainEngineKernel().LatestBlock()
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

// Get block summary information
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
	// Query block information

	store := api.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead()

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
		// Analyze miner information
		cbtx, _, e := transactions.ParseTransaction(blkbytes, coinbase_start_pos)
		if e != nil {
			result["err"] = e.Error()
			return result
		}
		// return
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
	// return
	result["jsondata"] = `{"datas":[` + strings.Join(jsondata, ",") + `]}`
	return result
}

// get miner message
func (api *DeprecatedApiService) getBlockMinerMsgStatistics(params map[string]string) map[string]string {
	result := make(map[string]string)

	last, _ := params["last_height"]
	step, _ := params["step_num"]
	limit, _ := params["limit"]
	last_hei, _ := strconv.ParseUint(last, 10, 0)
	if last_hei == 0 {
		result["err"] = "last_height param error"
		return result
	}
	step_num, _ := strconv.ParseUint(step, 10, 0)
	if step_num == 0 {
		step_num = 1
	}
	limit_num, _ := strconv.ParseUint(limit, 10, 0)
	if limit_num == 0 {
		limit_num = 1
	}

	store := api.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead()

	coinbase_start_pos := uint32(blocks.BlockHeadSize + blocks.BlockMetaSizeV1)
	coinbase_head_len := uint32(1 + 21 + 3 + 16 + 1)

	var minermsgary = []string{}

	var chei = last_hei
	for i := uint64(0); i < limit_num; i++ {

		blkhash, blkbytes, e := store.ReadBlockBytesByHeight(chei)
		if e != nil {
			result["err"] = e.Error()
			return result
		}
		if blkhash == nil || blkbytes == nil {
			result["err"] = "block height not find. " + fmt.Sprintf("", coinbase_head_len)
			return result
		}
		/*
			blkhead, _, e2 := blocks.ParseExcludeTransactions(blkbytes, 0)
			if e2 != nil {
				result["err"] = e2.Error()
				return result
			}
		*/
		// Analyze miner information
		cbtx, _, e := transactions.ParseTransaction(blkbytes, coinbase_start_pos)
		if e != nil {
			result["err"] = e.Error()
			return result
		}
		// return
		msg := cbtx.GetMessage().ValueShow()
		minermsgary = append(minermsgary, fmt.Sprintf(
			`{"hei":%d,"msg":"%s","adr":"%s"}`,
			chei, msg, cbtx.GetAddress().ToReadable(),
		))

		chei -= step_num
	}

	result["jsondata"] = `{"datas":[` + strings.Join(minermsgary, ",") + `]}`
	return result

}
