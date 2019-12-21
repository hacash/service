package rpc

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/memtxpool"
	"github.com/hacash/mint"
	"os"
	"strconv"

	"github.com/hacash/node/websocket"

	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	dealHomePrintCacheTime  = time.Now()
	dealHomePrintCacheBytes []byte
)

func (api *DeprecatedApiService) dealHome(response http.ResponseWriter, request *http.Request) {

	if len(dealHomePrintCacheBytes) > 0 && time.Now().Unix() < dealHomePrintCacheTime.Unix()+5 {
		response.Write(dealHomePrintCacheBytes)
		return
	}
	dealHomePrintCacheTime = time.Now()

	state := api.blockchain.State()
	//store := state.BlockStore()

	lastest, err := state.ReadLastestBlockHeadAndMeta()
	if err != nil {
		response.Write([]byte(err.Error()))
		return
	}

	// 矿工状态
	var responseStrAry = []string{}

	curheight := lastest.GetHeight()
	// 出块统计
	mint_num288dj := uint64(mint.AdjustTargetDifficultyNumberOfBlocks)
	mint_eachtime := mint.EachBlockRequiredTargetTime
	mint_eachtime_f := float32(mint_eachtime)

	prev288_90height := uint64(curheight) - (mint_num288dj * 30 * 3)
	prev288_30height := uint64(curheight) - (mint_num288dj * 30)
	prev288_7height := uint64(curheight) - (mint_num288dj * 7)
	prev288height := uint64(curheight) / mint_num288dj * mint_num288dj
	num288 := uint64(curheight) - prev288height
	if prev288_7height <= 0 {
		prev288_7height = 1
	}
	if prev288height <= 0 {
		prev288height = 1
	}

	lastestdiamond, err := state.ReadLastestDiamond()
	if err != nil {
		response.Write([]byte(err.Error()))
		return
	}

	diamondNumber := 0
	if lastestdiamond != nil {
		diamondNumber = int(lastestdiamond.Number)
	}
	responseStrAry = append(responseStrAry, fmt.Sprintf(
		"height: %d, tx: %d, hash: %s, difficulty: %d, create_time: %s, diamond number: %d",
		curheight,
		lastest.GetTransactionCount()-1,
		hex.EncodeToString(lastest.Hash()),
		lastest.GetDifficulty(),
		time.Unix(int64(lastest.GetTimestamp()), 0).Format("2006/01/02 15:04:05"),
		diamondNumber,
	))

	cost288_90miao := api.getMiao(lastest, prev288_90height, mint_num288dj*90)
	cost288_30miao := api.getMiao(lastest, prev288_30height, mint_num288dj*30)
	cost288_7miao := api.getMiao(lastest, prev288_7height, mint_num288dj*7)
	cost288miao := api.getMiao(lastest, prev288height, num288)
	// fmt.Println(prev288height, num288, cost288miao)
	responseStrAry = append(responseStrAry, fmt.Sprintf(
		"block average time, last quarter: %s ( %ds/%ds = %.2f), month: %s ( %ds/%ds = %.4f), week: %s ( %ds/%ds = %.4f), last from %d+%d: %s ( %ds/%ds = %f)",
		time.Unix(int64(cost288_90miao), 0).Format("04:05"),
		cost288_90miao, mint_eachtime,
		(float32(cost288_90miao)/mint_eachtime_f),
		time.Unix(int64(cost288_30miao), 0).Format("04:05"),
		cost288_30miao, mint_eachtime,
		(float32(cost288_30miao)/mint_eachtime_f),
		time.Unix(int64(cost288_7miao), 0).Format("04:05"),
		cost288_7miao, mint_eachtime,
		(float32(cost288_7miao)/mint_eachtime_f),
		prev288height,
		num288,
		time.Unix(int64(cost288miao), 0).Format("04:05"),
		cost288miao, mint_eachtime,
		(float32(cost288miao)/mint_eachtime_f),
	))
	// 交易池信息
	txpool := api.txpool
	if pool, ok := txpool.(*memtxpool.MemTxPool); ok {
		diamonds := ""
		hd := pool.GetDiamondCreateTxGroup().Head
		for i := 0; i < 200; i++ {
			if hd != nil {
				if as := hd.GetTx().GetActions(); len(as) > 0 {
					if as[0].Kind() == 4 {
						if dia, ok := as[0].(*actions.Action_4_DiamondCreate); ok {
							if len(diamonds) > 0 {
								diamonds += "/" + string(dia.Diamond)
							} else {
								diamonds = string(dia.Diamond)
							}
						}
					}
				}
				hd = hd.GetNext()
			} else {
				break
			}
		}
		plcount, plsize := pool.GetTotalCount()
		responseStrAry = append(responseStrAry, fmt.Sprintf(
			"txpool length: %d, size: %fkb, diamond: %s",
			plcount,
			float64(plsize)/1024,
			diamonds,
		))
	}
	/*
		// 矿池信息
		if len(config.Config.MiningPool.StatisticsDir) > 0 {
			minerpool := miner2.GetGlobalInstanceMiningPool()
			responseStrAry = append(responseStrAry, fmt.Sprintf("miner pool connected client: %d", minerpool.StateData.ClientCount))
		}
	*/

	/*
		// 节点连接信息
		//p2pserver := p2p.GetGlobalInstanceP2PServer()
		nodeinfo := p2pserver.GetServer().NodeInfo()
		p2pobj := p2p.GetGlobalInstanceProtocolManager()
		peers := p2pobj.GetPeers().PeersWithoutTx([]byte{0})
		bestpeername := ""
		for _, pr := range peers {
			bestpeername += pr.Name() + ", "
		}
		responseStrAry = append(responseStrAry, fmt.Sprintf(
			"p2p peer name: %s, enode: %s, connected: %d, connect peers: %s",
			nodeinfo.Name,
			nodeinfo.Enode,
			len(peers),
			strings.TrimRight(bestpeername, ", "),
		))
	*/

	// Write
	responseStrAry = append(responseStrAry, "")
	dealHomePrintCacheBytes = []byte("<html>" + strings.Join(responseStrAry, "\n\n<br><br> ") + "</html>")
	response.Write(dealHomePrintCacheBytes)
}

func (api *DeprecatedApiService) getMiao(minerblkhead interfaces.Block, prev288height uint64, blknum uint64) uint64 {

	prevblocktimestamp, err := api.blockchain.ReadPrev288BlockTimestamp(prev288height + 1)
	if err != nil {
		return 0
	}
	costtotalmiao := minerblkhead.GetTimestamp() - prevblocktimestamp
	if blknum == 0 {
		blknum = 1 // fix bug
	}
	costmiao := costtotalmiao / blknum
	return costmiao
}

func (api *DeprecatedApiService) dealQuery(response http.ResponseWriter, request *http.Request) {
	params := parseRequestQuery(request)
	if _, ok := params["action"]; !ok {
		response.Write([]byte("must action"))
		return
	}

	// call controller
	routeQueryRequest(params["action"], params, response, request)

}

func (api *DeprecatedApiService) dealOperate(response http.ResponseWriter, request *http.Request) {

	bodybytes, e1 := ioutil.ReadAll(request.Body)
	if e1 != nil {
		response.Write([]byte("body error"))
		return
	}
	if len(bodybytes) < 4 {
		response.Write([]byte("body length less than 4"))
		return
	}
	api.routeOperateRequest(response, binary.BigEndian.Uint32(bodybytes[0:4]), bodybytes[4:])
}

func parseRequestQuery(request *http.Request) map[string]string {
	request.ParseForm()
	params := make(map[string]string, 0)
	for k, v := range request.Form {
		//fmt.Println("key:", k)
		//fmt.Println("val:", strings.Join(v, ""))
		params[k] = strings.Join(v, "")
	}
	return params
}

func (api *DeprecatedApiService) RunHttpRpcService(port int) {

	api.initRoutes()

	mux := http.NewServeMux()

	mux.Handle("/websocket", websocket.Handler(api.webSocketHandler))

	mux.HandleFunc("/", api.dealHome)           //设置访问的路由
	mux.HandleFunc("/query", api.dealQuery)     //设置访问的路由
	mux.HandleFunc("/operate", api.dealOperate) //设置访问的路由

	//http.HandleFunc("/minerpool", minerPoolStatisticsAutoTransfer)       //设置访问的路由
	//http.HandleFunc("/minerpool/transactions", minerPoolAllTransactions) //设置访问的路由
	//http.HandleFunc("/minerpool/statistics", minerPoolStatistics) //设置访问的路由

	portstr := strconv.Itoa(port)

	// 设置监听的端口
	server := &http.Server{
		Addr:    ":" + portstr,
		Handler: mux,
	}

	fmt.Println("[Deprecated Api Service] Http listen on port: " + portstr)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
			os.Exit(0)
		} else {
			fmt.Println("RunHttpRpcService on " + portstr)
		}
	}()
}
