package main

import (
	"fmt"
	"testing"
)

func Test_bbb1(t *testing.T) {
	i1 := NewBalanceRankingItem("15BJ2PknNChqGLZRxWG8UUFXG9odQccvko", 10)
	i2 := NewBalanceRankingItem("1P8yzSQVkjhu1ACX6jKMrbpdnBgcjc2ATK", 8)
	i3 := NewBalanceRankingItem("127717zvZWFjEghjEpyyRSnitEEbnMuuLn", 15)
	i4 := NewBalanceRankingItem("127717zvZWFjEghjEpyyRSnitEEbnMuuLn", 9)

	t1 := make([]*BalanceRankingItem, 0)
	t1 = UpdateBalanceRankingTable(t1, i1, 10)
	t1 = UpdateBalanceRankingTable(t1, i2, 10)
	t1 = UpdateBalanceRankingTable(t1, i3, 10)
	t1 = UpdateBalanceRankingTable(t1, i4, 10)

	fmt.Println(t1[0].BalanceUint64(), t1[1].BalanceUint64(), t1[2].BalanceUint64())
}
