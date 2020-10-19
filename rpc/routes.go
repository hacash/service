package rpc

func (api *RpcService) initRoutes() {

	api.queryRoutes["create_accounts"] = api.createAccounts
	api.queryRoutes["create_value_transfer_tx"] = api.createValueTransferTx

	api.queryRoutes["balance"] = api.balance
	api.queryRoutes["diamond"] = api.diamond

	api.queryRoutes["last_block_height"] = api.lastBlockHeight
	api.queryRoutes["block_intro"] = api.blockIntro

	api.queryRoutes["scan_value_transfers"] = api.scanTransfersOfTransactionByPosition

}
