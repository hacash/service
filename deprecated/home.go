package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/miner/memtxpool"
	"github.com/hacash/mint"
	"github.com/hacash/mint/difficulty"
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

	kernel := api.blockchain.GetChainEngineKernel()
	//store := state.BlockStore()

	lastest, _, err := kernel.LatestBlock()
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

	prev288_365height := uint64(curheight) - (mint_num288dj * 30 * 12)
	prev288_90height := uint64(curheight) - (mint_num288dj * 30 * 3)
	prev288_30height := uint64(curheight) - (mint_num288dj * 30)
	prev288_7height := uint64(curheight) - (mint_num288dj * 7)
	prev288height := uint64(curheight) / mint_num288dj * mint_num288dj
	num288 := uint64(curheight) - prev288height
	if prev288_365height <= 0 {
		prev288_365height = 1
	}
	if prev288_90height <= 0 {
		prev288_90height = 1
	}
	if prev288_30height <= 0 {
		prev288_30height = 1
	}
	if prev288_7height <= 0 {
		prev288_7height = 1
	}
	if prev288height <= 0 {
		prev288height = 1
	}

	lastestdiamond, err := kernel.LatestDiamond()
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
		lastest.GetCustomerTransactionCount(),
		hex.EncodeToString(lastest.Hash()),
		lastest.GetDifficulty(),
		time.Unix(int64(lastest.GetTimestamp()), 0).Format("2006/01/02 15:04:05"),
		diamondNumber,
	))

	// powpower
	powpowerres := api.powPower(nil)
	if _, ok := powpowerres["err"]; !ok {
		responseStrAry = append(responseStrAry, fmt.Sprintf("real time pow power: %s", powpowerres["show"]))
	}

	//
	cost288_365miao := api.getMiao(lastest, prev288_365height, mint_num288dj*30*12)
	cost288_90miao := api.getMiao(lastest, prev288_90height, mint_num288dj*30*3)
	cost288_30miao := api.getMiao(lastest, prev288_30height, mint_num288dj*30)
	cost288_7miao := api.getMiao(lastest, prev288_7height, mint_num288dj*7)
	cost288miao := api.getMiao(lastest, prev288height, num288)
	// fmt.Println(prev288height, num288, cost288miao)
	responseStrAry = append(responseStrAry, fmt.Sprintf(
		"block average time, last year: %s ( %ds/%ds = %.2f), last quarter: %s ( %ds/%ds = %.2f), month: %s ( %ds/%ds = %.4f), week: %s ( %ds/%ds = %.4f), last from %d+%d: %s ( %ds/%ds = %f)",
		time.Unix(int64(cost288_365miao), 0).Format("04:05"),
		cost288_365miao, mint_eachtime,
		(float32(cost288_365miao)/mint_eachtime_f),
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
				if as := hd.GetTx().GetActionList(); len(as) > 0 {
					if as[0].Kind() == 4 {
						if dia, ok := as[0].(*actions.Action_4_DiamondCreate); ok {
							if len(diamonds) > 0 {
								diamonds += " / " + string(dia.Diamond)
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

	if api.backend != nil {
		responseStrAry = append(responseStrAry, api.backend.AllPeersDescribe())
	}

	// Write
	responseStrAry = append(responseStrAry, "")
	dealHomePrintCacheBytes = []byte("<html>" + strings.Join(responseStrAry, "\n\n<br><br> ") + "</html>")
	response.Write(dealHomePrintCacheBytes)
}

func (api *DeprecatedApiService) getMiao(minerblkhead interfaces.BlockHeadMetaRead, prev288height uint64, blknum uint64) uint64 {

	prevblocktimestamp, err := difficulty.ReadPrev288BlockTimestamp(api.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead(), prev288height+1)
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
