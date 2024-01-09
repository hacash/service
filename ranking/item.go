package main

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/core/fields"
	"math"
	"time"
)

type TransferTurnoverStatistic struct {
	WeekNum           fields.VarUint4
	trsCountHAC_float float64
	TrsCountHAC       fields.Bytes8
	TrsCountSAT       fields.VarUint8
	TrsCountHACD      fields.VarUint4
	UpdateTime        time.Time
	SaveedTime        time.Time
}

func NewTransferTurnoverStatistic() *TransferTurnoverStatistic {
	return &TransferTurnoverStatistic{
		WeekNum:           0,
		trsCountHAC_float: 0,
		TrsCountHAC:       make([]byte, 8),
		TrsCountSAT:       0,
		TrsCountHACD:      0,
		UpdateTime:        time.Now(),
		SaveedTime:        time.Now(),
	}
}

func (t *TransferTurnoverStatistic) AppendHAC(mei float64) {
	//fmt.Printf("********mei %f", mei)
	t.trsCountHAC_float += mei
}
func (t *TransferTurnoverStatistic) AppendSAT(sat uint64) {
	t.TrsCountSAT = fields.VarUint8(uint64(t.TrsCountSAT) + sat)
}
func (t *TransferTurnoverStatistic) AppendHACD(hacd uint32) {
	t.TrsCountHACD = fields.VarUint4(uint32(t.TrsCountHACD) + hacd)
}

func (t TransferTurnoverStatistic) GetHAC() float64 {
	return t.trsCountHAC_float
}
func (t TransferTurnoverStatistic) GetBTC() float64 {
	return float64(t.TrsCountSAT) / float64(10000_0000)
}
func (t TransferTurnoverStatistic) GetHACD() float64 {
	return float64(t.TrsCountHACD)
}

func (t *TransferTurnoverStatistic) Parse(buf []byte, seek uint32) (uint32, error) {
	var e error = nil
	seek, e = t.WeekNum.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = t.TrsCountHAC.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	var hacnum = math.Float64frombits(binary.BigEndian.Uint64(t.TrsCountHAC))
	t.trsCountHAC_float = hacnum
	seek, e = t.TrsCountSAT.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	seek, e = t.TrsCountHACD.Parse(buf, seek)
	if e != nil {
		return 0, e
	}
	return seek, nil
}

func (t *TransferTurnoverStatistic) Serialize() []byte {
	var bits = math.Float64bits(t.trsCountHAC_float)
	binary.BigEndian.PutUint64(t.TrsCountHAC, bits)
	//
	var buf = bytes.NewBuffer(nil)
	v1, _ := t.WeekNum.Serialize()
	v2, _ := t.TrsCountHAC.Serialize()
	v3, _ := t.TrsCountSAT.Serialize()
	v4, _ := t.TrsCountHACD.Serialize()
	buf.Write(v1)
	buf.Write(v2)
	buf.Write(v3)
	buf.Write(v4)
	// ok
	return buf.Bytes()
}

/******************************/

type BalanceRankingItem struct {
	Address       fields.Address
	BlsUseFloat64 fields.Bool // Whether the tag is a floating point number
	Balance       fields.VarUint8
}

func NewBalanceRankingItem(addrstr string, isfloat bool) *BalanceRankingItem {
	addr, _ := fields.CheckReadableAddress(addrstr)
	isf := uint8(0)
	if isfloat {
		isf = 1
	}

	return &BalanceRankingItem{
		Address:       *addr,
		BlsUseFloat64: fields.Bool(isf),
		Balance:       fields.VarUint8(0),
	}
}

func (b *BalanceRankingItem) GetBalance() float64 {
	if b.BlsUseFloat64.Check() {
		return math.Float64frombits(uint64(b.Balance))
	}
	return float64(b.Balance)
}

func (b *BalanceRankingItem) GetBalanceForceUint64() uint64 {
	return uint64(b.Balance)
}

func (b *BalanceRankingItem) SetBalanceByUint64(v uint64) {
	b.BlsUseFloat64.Set(false)
	b.Balance = fields.VarUint8(v)
}

func (b *BalanceRankingItem) SetBalanceByFloat64(v float64) {
	b.BlsUseFloat64.Set(true)
	uv := math.Float64bits(v)
	b.Balance = fields.VarUint8(uv)
}

// serialize
func ParseBalanceRankingItems(buf []byte) []*BalanceRankingItem {
	newtable := []*BalanceRankingItem{}
	blen := len(buf)
	ist := 21 + 1 + 8
	for seek := 0; seek < blen; {
		one := BalanceRankingItem{}
		if blen < seek+ist {
			break
		}
		one.Address = buf[seek : seek+21]
		seek += 21
		one.BlsUseFloat64 = fields.Bool(buf[seek])
		seek += 1
		one.Balance = fields.VarUint8(binary.BigEndian.Uint64(buf[seek : seek+8]))
		seek += 8
		newtable = append(newtable, &one)
	}

	return newtable
}

// Deserialization
func SerializeBalanceRankingItems(table []*BalanceRankingItem) []byte {
	buf := bytes.NewBuffer(nil)
	for _, v := range table {
		b1, _ := v.Address.Serialize()
		b2, _ := v.BlsUseFloat64.Serialize()
		b3, _ := v.Balance.Serialize()
		buf.Write(b1)
		buf.Write(b2)
		buf.Write(b3)
	}

	return buf.Bytes()
}

// Update ranking table
func UpdateBalanceRankingTable(table []*BalanceRankingItem, insert *BalanceRankingItem, maxsize int) []*BalanceRankingItem {
	istvzore := insert.GetBalance() == 0
	tlen := len(table)
	if tlen == 0 && !istvzore {
		return []*BalanceRankingItem{insert}
	}

	// duplicate removal
	if len(table) == 1 && table[0].Address.Equal(insert.Address) {
		return []*BalanceRankingItem{insert} // replace
	}

	var newtable []*BalanceRankingItem = nil
	for i := 0; i < tlen; i++ {
		li := table[i]
		if li.Address.Equal(insert.Address) {
			// Remove the current duplicate
			newtable = []*BalanceRankingItem{}
			newtable = append(newtable, table[0:i]...)
			newtable = append(newtable, table[i+1:]...)
			break
		}
	}

	if newtable != nil {
		table = newtable
	}

	// If the value is zero, delete it directly
	if istvzore {
		return table
	}

	// insert
	tlen = len(table)
	istidx := int(-1)
	b1 := insert.GetBalance()
	for i := tlen - 1; i >= 0; i-- {
		li := table[i]
		b2 := li.GetBalance()
		if b1 <= b2 {
			istidx = i
			break
		}
		if b1 > b2 {
			continue // Continue up
		}
	}

	// insert
	newtable = []*BalanceRankingItem{}
	if istidx == tlen-1 {
		// Add to end
		newtable = append(newtable, table...)
		newtable = append(newtable, insert)
	} else if istidx == -1 {
		// Add to first
		newtable = append(newtable, insert)
		newtable = append(newtable, table...)
	} else {
		// Insert middle
		newtable = append(newtable, table[0:istidx+1]...)
		newtable = append(newtable, insert)
		newtable = append(newtable, table[istidx+1:]...)
	}

	// Judge size
	if len(newtable) > maxsize {
		newtable = newtable[0:maxsize]
	}

	return newtable
}
