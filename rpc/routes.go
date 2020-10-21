package rpc

func (api *RpcService) initRoutes() {

	// submit

	// create
	api.createRoutes["accounts"] = api.createAccounts
	api.createRoutes["value_transfer_tx"] = api.createValueTransferTx

	// query
	api.queryRoutes["balances"] = api.balance
	api.queryRoutes["diamonds"] = api.diamond

	api.queryRoutes["last_block_height"] = api.lastBlockHeight
	api.queryRoutes["block_intro"] = api.blockIntro

	api.queryRoutes["scan_value_transfers"] = api.scanTransfersOfTransactionByPosition

}
