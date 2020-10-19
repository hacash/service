package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/stores"
	"github.com/hacash/x16rs"
	"net/http"
	"strings"
)

// 查询钻石
func (api *RpcService) diamond(r *http.Request, w http.ResponseWriter) {

	// 钻石 name or number
	diamondValue := strings.Trim(CheckParamString(r, "name", ""), " ")
	if len(diamondValue) > 0 {
		if len(diamondValue) != 6 {
			ResponseErrorString(w, "param diamond format error")
			return
		}
		if !x16rs.IsDiamondValueString(diamondValue) {
			ResponseError(w, fmt.Errorf("param diamond <%s> is not diamond name", diamondValue))
			return
		}
	}
	diamondNumber := CheckParamUint64(r, "number", 0)

	// read store
	var blockstore = api.backend.BlockChain().State().BlockStore()
	var err error
	var diamondSto *stores.DiamondSmelt = nil
	if diamondNumber > 0 {
		diamondSto, err = blockstore.ReadDiamondByNumber(uint32(diamondNumber))
		if err != nil {
			ResponseError(w, err)
			return
		}
	} else if len(diamondValue) == 6 {
		diamondSto, err = blockstore.ReadDiamond(fields.Bytes6(diamondValue))
		if err != nil {
			ResponseError(w, err)
			return
		}
	} else {
		ResponseErrorString(w, "param name or number must give")
		return
	}

	if diamondSto == nil {
		ResponseError(w, fmt.Errorf("diamond not find"))
		return
	}

	diamondValue = string(diamondSto.Diamond)

	// data
	retdata := ResponseCreateData("number", diamondSto.Number)
	retdata["name"] = diamondValue
	retdata["miner_address"] = diamondSto.MinerAddress.ToReadable()
	retdata["nonce"] = hex.EncodeToString(diamondSto.Nonce)
	retdata["custom_message"] = hex.EncodeToString(diamondSto.CustomMessage)
	retdata["contain_block_height"] = diamondSto.ContainBlockHeight
	retdata["contain_block_hash"] = hex.EncodeToString(diamondSto.ContainBlockHash)
	retdata["prev_block_hash"] = hex.EncodeToString(diamondSto.PrevContainBlockHash)

	// get current belong
	sto2 := api.backend.BlockChain().State().Diamond(fields.Bytes6(diamondValue))
	if sto2 != nil {
		retdata["current_belong_address"] = sto2.Address.ToReadable()
	} else {
		retdata["current_belong_address"] = diamondSto.MinerAddress.ToReadable()
	}

	// return
	ResponseData(w, retdata)
	return
}
