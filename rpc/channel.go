package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/stores"
	"net/http"
	"strings"
)

// 查询通道
func (api *RpcService) channel(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	// 是否以枚为单位
	isUnitMei := CheckParamBool(r, "unitmei", false)

	// 通道ID
	channelIdStr := strings.Trim(CheckParamString(r, "id", ""), " ")
	channelId, e := hex.DecodeString(channelIdStr)
	if e != nil || len(channelId) != stores.ChannelIdLength {
		ResponseErrorString(w, "channel id format error")
		return
	}

	// read store
	var blockstate = api.backend.BlockChain().State()
	//var err error

	channel := blockstate.Channel(channelId)
	if channel == nil {
		ResponseErrorString(w, "channel not find")
		return
	}

	// data
	retdata := ResponseCreateData("id", channelIdStr)
	retdata["status"] = channel.Status
	retdata["left_address"] = channel.LeftAddress.ToReadable()
	retdata["left_amount"] = channel.LeftAmount.ToMeiOrFinString(isUnitMei)
	retdata["right_address"] = channel.RightAddress.ToReadable()
	retdata["right_amount"] = channel.RightAmount.ToMeiOrFinString(isUnitMei)
	retdata["reuse_version"] = channel.ReuseVersion
	retdata["lock_block"] = channel.LockBlock

	// return
	ResponseData(w, retdata)
	return
}
