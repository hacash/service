package rpc

func (api *RpcService) initRoutes() {

	// submit
	api.submitRoutes["transaction"] = api.submitTransaction

	// create
	api.createRoutes["accounts"] = api.createAccounts
	api.createRoutes["value_transfer_tx"] = api.createValueTransferTx

	// query
	api.queryRoutes["balances"] = api.balances
	api.queryRoutes["diamond"] = api.diamond

	api.queryRoutes["last_block"] = api.lastBlock
	api.queryRoutes["block_intro"] = api.blockIntro

	api.queryRoutes["scan_value_transfers"] = api.scanTransfersOfTransactionByPosition

	// operate
	api.operateRoutes["raise_tx_fee"] = api.raiseTxFee

}
