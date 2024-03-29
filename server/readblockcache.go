package rpc

import (
	"fmt"
	"github.com/hacash/core/blocks"
	"github.com/hacash/core/interfaces"
	"sync"
)

/**
 * 读取区块数据的缓存
 */

const (
	rpcReadBlockCacheMaxSize int = 4 // Number of cache blocks
)

var (
	rpcReadBlockCacheDatas = make([]interfaces.Block, 0)
	rpcReadBlockCacheMux   sync.Mutex
)

// Load block
func LoadBlockWithCache(kernel interfaces.ChainEngine, height uint64) (interfaces.Block, error) {
	rpcReadBlockCacheMux.Lock()
	defer rpcReadBlockCacheMux.Unlock()

	// check cache
	for _, v := range rpcReadBlockCacheDatas {
		if height == v.GetHeight() {
			return v, nil // return cache
		}
	}

	// load from disk
	last, _, e1 := kernel.LatestBlock()
	if e1 != nil {
		return nil, e1
	}

	if height > last.GetHeight() {
		return nil, fmt.Errorf("block is not find.")
	}

	_, blkbody, err := kernel.StateRead().BlockStoreRead().ReadBlockBytesByHeight(height)
	if err != nil {
		return nil, err
	}

	blockObj, _, err2 := blocks.ParseBlock(blkbody, 0)
	if err2 != nil {
		return nil, err2
	}

	// put in cache
	rpcReadBlockCacheDatas = append(rpcReadBlockCacheDatas, blockObj)
	// check size
	if len(rpcReadBlockCacheDatas) > rpcReadBlockCacheMaxSize {
		rpcReadBlockCacheDatas = rpcReadBlockCacheDatas[1:]
	}

	// read ok
	return blockObj, nil
}
