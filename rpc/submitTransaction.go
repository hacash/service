package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/transactions"
	"net/http"
)

// 提交一笔交易
func (api *RpcService) submitTransaction(r *http.Request, w http.ResponseWriter) {

	txbody := r.PostFormValue("txbody")
	if len(txbody) == 0 {
		ResponseErrorString(w, "post data <txbody> must give")
		return
	}
	bodybytes := []byte(txbody)
	//fmt.Println(len(bodybytes), string(bodybytes))

	isHexData := CheckParamBool(r, "hexbody", false)

	// hex字符串方式
	if isHexData {
		realbodybts, e2 := hex.DecodeString(string(bodybytes))
		if e2 != nil {
			ResponseError(w, e2)
			return
		}
		bodybytes = realbodybts
	}

	// 解析交易
	trs, _, e3 := transactions.ParseTransaction(bodybytes, 0)
	if e3 != nil {
		ResponseError(w, e3)
		return
	}

	// 尝试加入交易池
	e4 := api.txpool.AddTx(trs)
	if e4 != nil {
		ResponseError(w, e4)
		return
	}

	// 返回成功
	ResponseData(w, nil)
	return
}
