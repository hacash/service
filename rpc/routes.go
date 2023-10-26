package rpc

func (api *RpcService) initRoutes() {
	// submit
	api.submitRoutes["transaction"] = api.submitTransaction

	// create
	api.createRoutes["accounts"] = api.createAccounts
	api.createRoutes["value_transfer_tx"] = api.createValueTransferTx

	// query
	api.queryRoutes["total_supply"] = api.totalSupply
	api.queryRoutes["balances"] = api.balances
	api.queryRoutes["diamond"] = api.diamond
	api.queryRoutes["channel"] = api.channel
	api.queryRoutes["last_block"] = api.lastBlock
	api.queryRoutes["block_intro"] = api.blockIntro
	api.queryRoutes["scan_value_transfers"] = api.scanTransfersOfTransactionByPosition
	api.queryRoutes["scan_coin_transfers"] = api.scanCoinTransfersOfTransactionByPosition
	api.queryRoutes["hdns"] = api.hdns

	// operate
	api.operateRoutes["raise_tx_fee"] = api.raiseTxFee
}
