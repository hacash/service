package main

import "time"

func (r *Ranking) loop() {

	for {

		tickerFlushStateToDisk := time.NewTicker(time.Minute * time.Duration(r.flush_state_timeout_minute))
		tickerUploadTotalSupply := time.NewTicker(time.Minute * 23)

		select {
		case <-r.flushStateToDiskNotifyCh:
			go r.flushStateToDisk() // 通知更新储存
		case <-tickerFlushStateToDisk.C:
			go r.flushStateToDisk() // 定期更新储存
		case <-tickerUploadTotalSupply.C:
			go r.loadTotalSupply() // 定期更新最大供应量
		}
	}

}
