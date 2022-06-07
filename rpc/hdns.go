package rpc

import (
	"fmt"
	"github.com/hacash/core/fields"
	"github.com/hacash/x16rs"
	"net/http"
	"strconv"
)

// Hdns resolution service
func (api *RpcService) hdns(r *http.Request, w http.ResponseWriter, bodybytes []byte) {
	// Diamond face value or number
	diamondStr := CheckParamString(r, "diamond", "")
	if len(diamondStr) == 0 {
		ResponseErrorString(w, "param diamond must give")
		return
	}

	// number
	if dianum, e := strconv.Atoi(diamondStr); e == nil && dianum > 0 && dianum < 16777216 {
		disobj, e := api.backend.BlockChain().GetChainEngineKernel().StateRead().BlockStoreRead().ReadDiamondByNumber(uint32(dianum))
		if e == nil || disobj != nil {
			diamondStr = string(disobj.Diamond) // Number mapping to literal value
		}
	}

	// Literal 
	if x16rs.IsDiamondValueString(diamondStr) {
		disobj, _ := api.backend.BlockChain().GetChainEngineKernel().StateRead().Diamond(fields.DiamondName(diamondStr))
		if disobj != nil {
			data := map[string]interface{}{
				"address": disobj.Address.ToReadable(),
			}
			ResponseData(w, data)
			return // Resolution succeeded
		}
	}

	// No diamonds found
	ResponseErrorString(w, fmt.Sprintf("diamond <%s> not find.", diamondStr))
	return
}
