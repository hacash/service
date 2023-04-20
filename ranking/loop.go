package main

import (
	"time"
)

func (r *Ranking) loop() {

	var tickerFlushStateToDisk = time.NewTicker(time.Minute * time.Duration(r.flush_state_timeout_minute))
	var tickerUploadTotalSupply = time.NewTicker(time.Minute * 23)
	var tickerFlushTransferTurnoverToDisk = time.NewTicker(time.Second * 5)

	for {
		select {
		case <-tickerFlushTransferTurnoverToDisk.C:
			var turn = r.cache_turnover_curobj
			if turn.UpdateTime != turn.SaveedTime {
				turn.SaveedTime = turn.UpdateTime
				//fmt.Printf("flush Transfer Turnover %d\n", turn.WeekNum)
				go r.flushTransferTurnover(turn)
			}
		case <-r.flushStateToDiskNotifyCh:
			go r.flushStateToDisk() // Notification update storage
		case <-tickerFlushStateToDisk.C:
			go r.flushStateToDisk() // Regular update storage
		case <-tickerUploadTotalSupply.C:
			go r.loadTotalSupply() // Update the maximum supply regularly
		}
	}
}
