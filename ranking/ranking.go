package main

import (
	"fmt"
	"github.com/hacash/chain/leveldb"
	"github.com/hacash/core/sys"
	"os"
	"sync"
)

const (
	DBKeyFinishScanBlockHeight    = "finish_scan_block_height"
	DBKeyHacashBalanceRanking100  = "hacash_balance_ranking_100"
	DBKeyDiamondBalanceRanking100 = "diamond_balance_ranking_100"
	DBKeySatoshiBalanceRanking100 = "satoshi_balance_ranking_100"
)

type Ranking struct {

	// config
	ldb_dir               string
	http_api_listen_port  int    // api 接口监听
	balance_ranking_range int    // 余额排名范围
	node_rpc_url          string // 节点rpc接口地址： http://127.0.0.1:8083

	// ptr
	ldb                      *leveldb.DB
	dataChangeLocker         sync.Mutex
	flushStateToDiskNotifyCh chan bool // 保存数据通知

	// data
	finish_scan_block_height    uint64                // 已经扫描完成的区块id
	hacash_balance_ranking_100  []*BalanceRankingItem // HAC 持仓前100
	diamond_balance_ranking_100 []*BalanceRankingItem // HACD 持仓前100
	satoshi_balance_ranking_100 []*BalanceRankingItem // BTC 持仓前100

	wait_update_address_num  int               // 等待更新的 100 个地址
	wait_update_address_list map[string]bool   // 等待更新的 100 个地址
	cache_update_diamonds    map[string][]byte // 等待更新的地址和钻石表

	// current_circulation
	current_circulation float64
	minted_diamond      int64
	transferred_bitcoin int64
}

func NewRanking(cnffile *sys.Inicnf) *Ranking {

	section := cnffile.Section("")

	ldb_dir := section.Key("data_dir").MustString("./hacash_ranking_data")
	apiport := section.Key("http_api_listen_port").MustInt(3377)
	rpc_api := section.Key("node_rpc_ip_port").MustString("http://127.0.0.1:8083")

	// create
	rank := &Ranking{
		http_api_listen_port: apiport,
		ldb_dir:              ldb_dir,
		node_rpc_url:         rpc_api,

		balance_ranking_range:       100, // 持仓排名100
		finish_scan_block_height:    0,
		flushStateToDiskNotifyCh:    make(chan bool, 10), // 通知管道
		hacash_balance_ranking_100:  make([]*BalanceRankingItem, 0),
		diamond_balance_ranking_100: make([]*BalanceRankingItem, 0),
		satoshi_balance_ranking_100: make([]*BalanceRankingItem, 0),
		wait_update_address_num:     0,
		wait_update_address_list:    make(map[string]bool),
		cache_update_diamonds:       make(map[string][]byte),
	}

	return rank
}

func (r *Ranking) Start() {
	var e error

	dbdir := sys.AbsDir(r.ldb_dir)
	r.ldb, e = leveldb.OpenFile(dbdir, nil)
	if e != nil {
		fmt.Println("[ERROR] open leveldb error: ", e)
		os.Exit(0)
	}

	// 检查rpc
	_, err := HttpGetBytes(r.node_rpc_url)
	if err != nil {
		fmt.Println("[ERROR] check node_rpc_url error: ", err)
		os.Exit(0)
	}

	e2 := r.init()
	if e2 != nil {
		fmt.Println("[ERROR] init error: ", e2)
		os.Exit(0)
	}

	go r.startHttpApiService()

	// loop
	go r.loop()

	go r.scan()

}
