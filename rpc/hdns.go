package rpc

import (
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/x16rs"
	"net/http"
	"strconv"
)

// HDNS 解析服务
func (api *RpcService) hdns(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	// 钻石字面值或编号
	diamondStr := CheckParamString(r, "diamond", "")
	if len(diamondStr) == 0 {
		ResponseErrorString(w, "param diamond must give")
		return
	}

	// 编号
	if dianum, e := strconv.Atoi(diamondStr); e == nil && dianum > 0 && dianum < 16777216 {
		disobj, e := api.backend.BlockChain().StateRead().BlockStoreRead().ReadDiamondByNumber(uint32(dianum))
		if e == nil || disobj != nil {
			diamondStr = string(disobj.Diamond) // 编号映射到字面值
		}
	}

	// 字面值
	if x16rs.IsDiamondValueString(diamondStr) {
		disobj, _ := api.backend.BlockChain().StateRead().Diamond(fields.DiamondName(diamondStr))
		if disobj != nil {
			data := map[string]interface{}{
				"address": disobj.Address.ToReadable(),
			}
			ResponseData(w, data)
			return // 解析成功
		}
	}

	// 没找到钻石
	ResponseErrorString(w, fmt.Sprintf("diamond <%s> not find.", diamondStr))
	return
}
