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
	ldb_dir                    string
	http_api_listen_port       int    // API interface listening
	balance_ranking_range      int    // Balance ranking range
	node_rpc_url               string // Node RPC interface address: http://127.0.0.1:8083
	flush_state_timeout_minute int    // Refresh balance interval (minutes)

	// ptr
	ldb                      *leveldb.DB
	dataChangeLocker         sync.Mutex
	flushStateToDiskNotifyCh chan bool // Save data notification

	// data
	finish_scan_block_height    uint64                // Scanned block ID
	hacash_balance_ranking_100  []*BalanceRankingItem // Top 100 HAC positions
	diamond_balance_ranking_100 []*BalanceRankingItem // Top 100 hacd positions
	satoshi_balance_ranking_100 []*BalanceRankingItem // Top 100 BTC positions

	wait_update_address_num  int               // 100 addresses waiting to be updated
	wait_update_address_list map[string]bool   // 100 addresses waiting to be updated
	cache_update_diamonds    map[string][]byte // Address and diamond table waiting to be updated

	cache_turnover_curobj *TransferTurnoverStatistic

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
	flush_time := section.Key("flush_state_timeout_minute").MustInt(30)

	// create
	rank := &Ranking{
		http_api_listen_port: apiport,
		ldb_dir:              ldb_dir,
		node_rpc_url:         rpc_api,

		flush_state_timeout_minute:  flush_time,
		balance_ranking_range:       100, // 持仓排名100
		finish_scan_block_height:    0,
		flushStateToDiskNotifyCh:    make(chan bool, 10), // 通知管道
		hacash_balance_ranking_100:  make([]*BalanceRankingItem, 0),
		diamond_balance_ranking_100: make([]*BalanceRankingItem, 0),
		satoshi_balance_ranking_100: make([]*BalanceRankingItem, 0),
		wait_update_address_num:     0,
		wait_update_address_list:    make(map[string]bool),
		cache_update_diamonds:       make(map[string][]byte),
		cache_turnover_curobj:       NewTransferTurnoverStatistic(),
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

	// Check RPC
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
