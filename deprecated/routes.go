package rpc

import (
	//"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	queryRoutes = make(map[string]func(map[string]string) map[string]string)
)

func (api *DeprecatedApiService) initRoutes() {
	queryRoutes["balance"] = api.getBalance                 // 查询余额
	queryRoutes["diamond"] = api.getDiamond                 // 查询钻石
	queryRoutes["diamondcreate"] = api.showDiamondCreateTxs // 显示钻石创建交易
	queryRoutes["transferdiamonds"] = api.transferDiamondMultiple // 钻石批量转账
	queryRoutes["channel"] = api.getChannel                 // 查询通道
	queryRoutes["passwd"] = newAccountByPassword            // 通过密码创建账户
	queryRoutes["newacc"] = newAccount                      // 随机创建账户
	queryRoutes["createtx"] = api.transferSimple            // 创建普通转账交易
	queryRoutes["quotefee"] = api.quoteFee                  // 修改交易池内的交易手续费
	queryRoutes["txconfirm"] = api.txStatus                 // 查询交易确认状态
	queryRoutes["powpower"] = api.powPower                  // 实时算力大小

	queryRoutes["blocks"] = api.getBlockAbstractList  // 查询区块信息
	queryRoutes["lastblock"] = api.getLastBlockHeight // 查询最新区块高度
	queryRoutes["blockintro"] = api.getBlockIntro     // 查询区块简介
	queryRoutes["trsintro"] = api.getTransactionIntro // 查询区块简介

	queryRoutes["getalltransferlogbyblockheight"] = api.getAllTransferLogByBlockHeight // 扫描区块 获取所有转账信息

}

func routeQueryRequest(action string, params map[string]string, w http.ResponseWriter, r *http.Request) {
	if ctrl, ok := queryRoutes[action]; ok {
		resobj := ctrl(params)
		w.Header().Set("Content-Type", "text/json")
		if jsondata, ok := resobj["jsondata"]; ok {
			w.Write([]byte(jsondata)) // 自定义的 jsondata 数据
		} else {
			restxt, e1 := json.Marshal(resobj)
			if e1 != nil {
				w.Write([]byte("data not json"))
			} else {
				w.Write(restxt)
			}
		}
	} else {
		w.Write([]byte("not find action"))
	}
}

func (api *DeprecatedApiService) routeOperateRequest(w http.ResponseWriter, opcode uint32, value []byte) {
	switch opcode {
	/////////////////////////////
	case 1:
		api.addTxToPool(w, value)
	/////////////////////////////
	default:
		w.Write([]byte(fmt.Sprint("not find opcode %d", opcode)))
	}
}

//////////////////////////////////////////////////////////////