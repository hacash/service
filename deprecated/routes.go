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
	queryRoutes["balance"] = api.getBalance                       // Check the balance
	queryRoutes["diamond"] = api.getDiamond                       // Query diamond
	queryRoutes["diamondcreate"] = api.showDiamondCreateTxs       // Show diamond creation transactions
	queryRoutes["transferdiamonds"] = api.transferDiamondMultiple // Diamond batch transfer

	queryRoutes["channel"] = api.getChannel              // Query channel
	queryRoutes["lockbls"] = api.getLockBlsInfo          // Query linear lock
	queryRoutes["dialend"] = api.getSystemLendingDiamond // Query diamond system loan
	queryRoutes["btclend"] = api.getSystemLendingBitcoin // Query debit and credit of bitcoin system
	queryRoutes["usrlend"] = api.getUserLending          // Query debit and credit of bitcoin system

	queryRoutes["passwd"] = newAccountByPassword // Create account with password
	queryRoutes["newacc"] = newAccount           // Random account creation
	queryRoutes["createtx"] = api.transferSimple // Create a general transfer transaction
	queryRoutes["quotefee"] = api.quoteFee       // Modify the transaction service charge in the trading pool
	queryRoutes["txconfirm"] = api.txStatus      // Query transaction confirmation status
	queryRoutes["powpower"] = api.powPower       // Real time calculation force

	queryRoutes["hashrate"] = api.hashRate                   // Current hash rate
	queryRoutes["hashrate_charts"] = api.hashRateCharts      // Hash rate fluctuation table
	queryRoutes["hashrate_charts_v3"] = api.hashRateChartsV3 // Hash rate fluctuation table

	queryRoutes["blocks"] = api.getBlockAbstractList                   // Query block information
	queryRoutes["lastblock"] = api.getLastBlockHeight                  // Query the latest block height
	queryRoutes["blockintro"] = api.getBlockIntro                      // Query block introduction
	queryRoutes["blockdatahex"] = api.getBlockDataOfHex                // Query block body data
	queryRoutes["changeblockreferheight"] = api.changeBlockReferHeight // Change block height pointer
	queryRoutes["trsintro"] = api.getTransactionIntro                  // Query transaction introduction
	queryRoutes["recentblocks"] = api.getRecentArrivedBlockList        // Query recent arrived blocks

	queryRoutes["getalltransferlogbyblockheight"] = api.getAllTransferLogByBlockHeight           // Scan the block to obtain all transfer information
	queryRoutes["getalloperateactionlogbyblockheight"] = api.getAllOperateActionLogByBlockHeight // Scan the block to obtain operation logs other than transfer

	queryRoutes["getdiamondvisualgenelist"] = api.getDiamondVisualGeneList // Get diamond appearance gene list

	queryRoutes["btcmovelog"] = api.getBtcMoveLogPageData // Get bitcoin transfer log page data

	queryRoutes["totalsupply"] = api.totalSupply // Total supply
	queryRoutes["totalnonemptyaccount"] = api.getTotalNonEmptyAccount
	queryRoutes["execfee"] = api.getLatestAverageFeePurity

	// dex
	queryRoutes["dexbuycreate"] = api.dexBuyCreate     // Create a bill
	queryRoutes["dexbuyconfirm"] = api.dexBuyConfirm   // Confirm the bill
	queryRoutes["dexsellconfirm"] = api.dexSellConfirm // Confirm sales order
}

func routeQueryRequest(action string, params map[string]string, w http.ResponseWriter, r *http.Request) {
	if ctrl, ok := queryRoutes[action]; ok {
		resobj := ctrl(params)
		w.Header().Set("Content-Type", "text/json;charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if jsondata, ok := resobj["jsondata"]; ok {
			w.Write([]byte(jsondata)) // Customized jsondata data
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	switch opcode {
	/////////////////////////////
	case 1:
		api.addTxToPool(w, value)
	case 2:
		api.createTxAndCheckOrCommit(w, value)
	/////////////////////////////
	default:
		w.Write([]byte(fmt.Sprint("not find opcode %d", opcode)))
	}
}
