package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"net/http"
)

// 提交一笔交易
func (api *RpcService) submitTransaction(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	isHexData := CheckParamBool(r, "hexbody", false)

	// hex字符串方式
	if isHexData {
		realbodybts, e2 := hex.DecodeString(string(bodybytes))
		if e2 != nil {
			ResponseError(w, e2)
			return
		}
		//fmt.Println(len(realbodybts))
		//fmt.Println(string(realbodybts))
		bodybytes = realbodybts
	}

	// 解析交易
	trs, _, e3 := transactions.ParseTransaction(bodybytes, 0)
	if e3 != nil {
		ResponseError(w, e3)
		return
	}

	// 尝试加入交易池
	e4 := api.txpool.AddTx(trs.(interfaces.Transaction))
	if e4 != nil {
		ResponseError(w, e4)
		return
	}

	// 返回成功
	// return: status = success
	ResponseData(w, ResponseCreateData("status", "success"))

	return
}
