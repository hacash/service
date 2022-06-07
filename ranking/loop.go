package main

import "time"

func (r *Ranking) loop() {
	for {
		tickerFlushStateToDisk := time.NewTicker(time.Minute * time.Duration(r.flush_state_timeout_minute))
		tickerUploadTotalSupply := time.NewTicker(time.Minute * 23)

		select {
		case <-r.flushStateToDiskNotifyCh:
			go r.flushStateToDisk() // Notification update storage
		case <-tickerFlushStateToDisk.C:
			go r.flushStateToDisk() // Regular update storage
		case <-tickerUploadTotalSupply.C:
			go r.loadTotalSupply() // Update the maximum supply regularly
		}
	}
}
