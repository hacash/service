package main

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/core/fields"
	"math"
)

type BalanceRankingItem struct {
	Address         fields.Address
	Balance         fields.VarUint8
	bls_use_float64 bool
}

func NewBalanceRankingItem(addrstr string, bls uint64) *BalanceRankingItem {
	addr, _ := fields.CheckReadableAddress(addrstr)
	return &BalanceRankingItem{
		Address: *addr,
		Balance: fields.VarUint8(bls),
	}
}
func (b *BalanceRankingItem) BalanceUint64() uint64 {
	return uint64(b.Balance)
}
func (b *BalanceRankingItem) BalanceFloat64() float64 {
	v := math.Float64frombits(uint64(b.Balance))
	return v
}
func (b *BalanceRankingItem) SetBalanceUint64(v uint64) {
	b.bls_use_float64 = false
	b.Balance = fields.VarUint8(v)
}
func (b *BalanceRankingItem) SetBalanceFloat64(v float64) {
	b.bls_use_float64 = true
	uv := math.Float64bits(v)
	b.Balance = fields.VarUint8(uv)
}

// 序列化
func ParseBalanceRankingItems(buf []byte) []*BalanceRankingItem {
	newtable := []*BalanceRankingItem{}
	blen := len(buf)
	ist := 21 + 8
	for seek := 0; seek < blen; {
		one := BalanceRankingItem{}
		if blen < seek+ist {
			break
		}
		one.Address = buf[seek : seek+21]
		one.Balance = fields.VarUint8(binary.BigEndian.Uint64(buf[seek+21 : seek+21+8]))
		seek += ist
		newtable = append(newtable, &one)
	}
	return newtable
}

// 反序列化
func SerializeBalanceRankingItems(table []*BalanceRankingItem) []byte {
	buf := bytes.NewBuffer(nil)
	for _, v := range table {
		b1, _ := v.Address.Serialize()
		b2, _ := v.Balance.Serialize()
		buf.Write(b1)
		buf.Write(b2)
	}
	return buf.Bytes()
}

// 更新排名表
func UpdateBalanceRankingTable(table []*BalanceRankingItem, insert *BalanceRankingItem, maxsize int) []*BalanceRankingItem {
	istvzore := (insert.bls_use_float64 && insert.BalanceFloat64() == 0) ||
		(!insert.bls_use_float64 && insert.BalanceUint64() == 0)
	tlen := len(table)
	if tlen == 0 && !istvzore {
		return []*BalanceRankingItem{insert}
	}
	// 去重
	if len(table) == 1 && table[0].Address.Equal(insert.Address) {
		return []*BalanceRankingItem{insert} // 替换
	}
	var newtable []*BalanceRankingItem = nil
	for i := 0; i < tlen; i++ {
		li := table[i]
		if li.Address.Equal(insert.Address) {
			// 去掉当前这一个重复的
			newtable = []*BalanceRankingItem{}
			newtable = append(newtable, table[0:i]...)
			newtable = append(newtable, table[i+1:]...)
			break
		}
	}
	if newtable != nil {
		table = newtable
	}
	// 如果值为零则直接删掉
	if istvzore {
		return table
	}
	// 插入
	tlen = len(table)
	istidx := int(-1)
	b1 := float64(insert.Balance)
	if insert.bls_use_float64 {
		b1 = math.Float64frombits(uint64(insert.Balance))
	}
	for i := tlen - 1; i >= 0; i-- {
		li := table[i]
		b2 := float64(li.Balance)
		if li.bls_use_float64 {
			b2 = math.Float64frombits(uint64(li.Balance))
		}
		if b1 <= b2 {
			istidx = i
			break
		}
		if b1 > b2 {
			continue // 继续向上
		}
	}
	// 插入
	newtable = []*BalanceRankingItem{}
	if istidx == tlen-1 {
		// 加到末尾
		newtable = append(newtable, table...)
		newtable = append(newtable, insert)
	} else if istidx == -1 {
		// 加到第一位
		newtable = append(newtable, insert)
		newtable = append(newtable, table...)
	} else {
		// 插入中间
		newtable = append(newtable, table[0:istidx+1]...)
		newtable = append(newtable, insert)
		newtable = append(newtable, table[istidx+1:]...)
	}

	// 判断大小
	if len(newtable) > maxsize {
		newtable = newtable[0:maxsize]
	}

	return newtable
}
