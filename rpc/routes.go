package rpc

func (api *RpcService) initRoutes() {

	api.queryRoutes["last_block_height"] = api.lastBlockHeight
	api.queryRoutes["block_intro"] = api.blockIntro

	api.queryRoutes["scan_value_transfers"] = api.scanTransfersOfTransactionByPosition

}
