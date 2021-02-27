package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/account"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/transactions"
	"github.com/hacash/x16rs"
	"net/http"
	"strings"
	"time"
)

// 创建价值转移交易
func (api *RpcService) createValueTransferTx(r *http.Request, w http.ResponseWriter, bodybytes []byte) {

	var err error

	// mei
	isUnitMei := CheckParamBool(r, "unitmei", false)

	// address
	mainPrikeyStr := strings.TrimPrefix(CheckParamString(r, "main_prikey", ""), "0x")
	mainPrikeyStr = strings.TrimPrefix(mainPrikeyStr, "0X")
	if len(mainPrikeyStr) == 0 {
		ResponseErrorString(w, "param main_prikey must give")
		return
	}
	if len(mainPrikeyStr) != 64 {
		ResponseErrorString(w, "param main_prikey length error")
		return
	}
	var prikeybts []byte
	if prikeybts, err = hex.DecodeString(mainPrikeyStr); err != nil {
		ResponseErrorString(w, "param main_prikey format error")
		return
	}
	mainAccount, err := account.GetAccountByPriviteKey(prikeybts)

	// fee
	feeStr := CheckParamString(r, "fee", "")
	if len(feeStr) == 0 {
		ResponseErrorString(w, "param fee must give")
		return
	}
	var feeAmt *fields.Amount = nil
	if isUnitMei {
		feeAmt, err = fields.NewAmountFromMeiString(feeStr)
	} else {
		feeAmt, err = fields.NewAmountFromFinString(feeStr)
	}
	if err != nil {
		ResponseErrorString(w, "param fee format error")
		return
	}

	// timestamp
	curt := uint64(time.Now().Unix())
	timestamp := CheckParamUint64(r, "timestamp", 0)
	if timestamp > 0 {
		if timestamp > curt {
			ResponseErrorString(w, "param timestamp cant not over now")
			return
		}
	} else {
		timestamp = curt // def now
	}

	// prikey
	allPrivateKeyBytes := make(map[string][]byte, 1)
	allPrivateKeyBytes[string(mainAccount.Address)] = mainAccount.PrivateKey

	// create tx
	newTrs, e1 := transactions.NewEmptyTransaction_2_Simple(mainAccount.Address)
	if e1 != nil {
		ResponseErrorString(w, "create tx error: "+e1.Error())
		return
	}
	newTrs.Timestamp = fields.VarUint5(timestamp)
	newTrs.Fee = *feeAmt

	// create action
	transferKind := CheckParamString(r, "transfer_kind", "")
	var kinderr error
	switch transferKind {
	case "hacash":
		// 转账 hac
		kinderr = appendActionSimpleTransferHacash(r, isUnitMei, allPrivateKeyBytes, mainAccount, newTrs)
	case "satoshi":
		// 转账 btc
		kinderr = appendActionSimpleTransferSatoshi(r, isUnitMei, allPrivateKeyBytes, mainAccount, newTrs)
	case "diamond":
		// 转账 diamond
		kinderr = appendActionTransferDiamond(r, isUnitMei, allPrivateKeyBytes, mainAccount, newTrs)
	default:
		kinderr = fmt.Errorf("not find transfer_kind <%s>", transferKind)
	}

	if kinderr != nil {
		ResponseError(w, kinderr)
		return
	}

	// sign
	e9 := newTrs.FillNeedSigns(allPrivateKeyBytes, nil)
	if e9 != nil {
		ResponseError(w, e9)
		return
	}

	// data
	data := ResponseCreateData("hash", newTrs.HashFresh().ToHex())
	txbody, e10 := newTrs.Serialize()
	if e10 != nil {
		ResponseError(w, e10)
		return
	}
	data["hash_with_fee"] = newTrs.HashWithFeeFresh().ToHex()
	data["body"] = hex.EncodeToString(txbody)
	data["timestamp"] = timestamp

	// return
	ResponseData(w, data)
	return
}

/////////////////////////////////////////////////////////////

/**
 * 创建一笔普通转账
 */
func appendActionSimpleTransferHacash(r *http.Request, isUnitMei bool, allprikey map[string][]byte, mainAccount *account.Account, tx *transactions.Transaction_2_Simple) error {
	var err error

	// amount
	amountStr := CheckParamString(r, "amount", "")
	if len(amountStr) == 0 {
		return fmt.Errorf("param amount must give")
	}
	var amountAmt *fields.Amount = nil
	if isUnitMei {
		amountAmt, err = fields.NewAmountFromMeiString(amountStr)
	} else {
		amountAmt, err = fields.NewAmountFromFinString(amountStr)
	}
	if err != nil {
		return fmt.Errorf("param amount format error")
	}

	// address
	addrStr := CheckParamString(r, "to_address", "")
	if len(addrStr) == 0 {
		return fmt.Errorf("param to_address must give")
	}
	to_addr, e8 := fields.CheckReadableAddress(addrStr)
	if e8 != nil {
		return e8
	}

	var actObj = &actions.Action_1_SimpleTransfer{
		ToAddress: *to_addr,
		Amount:    *amountAmt,
	}

	tx.AppendAction(actObj)

	// success
	return nil

}

/**
 * 创建一笔普通转 satoshi
 */
func appendActionSimpleTransferSatoshi(r *http.Request, isUnitMei bool, allprikey map[string][]byte, mainAccount *account.Account, tx *transactions.Transaction_2_Simple) error {
	// amount
	amountAmt := CheckParamUint64(r, "amount", 0)
	if amountAmt == 0 {
		return fmt.Errorf("param amount error")
	}

	// address
	addrStr := CheckParamString(r, "to_address", "")
	if len(addrStr) == 0 {
		return fmt.Errorf("param to_address must give")
	}
	to_addr, e8 := fields.CheckReadableAddress(addrStr)
	if e8 != nil {
		return e8
	}

	var actObj = &actions.Action_8_SimpleSatoshiTransfer{
		Address: *to_addr,
		Amount:  fields.VarUint8(amountAmt),
	}

	tx.AppendAction(actObj)

	// success
	return nil

}

/**
 * 创建一笔钻石转账
 */
func appendActionTransferDiamond(r *http.Request, isUnitMei bool, allprikey map[string][]byte, mainAccount *account.Account, tx *transactions.Transaction_2_Simple) error {

	var err error
	var isMultiTrs bool = false

	// diamonds
	diamondsStr := strings.Trim(CheckParamString(r, "diamonds", ""), " ")
	if len(diamondsStr) == 0 {
		return fmt.Errorf("param diamonds must give")
	}
	diamonds := strings.Split(diamondsStr, ",")
	if len(diamonds) > 1 {
		isMultiTrs = true
	}
	if len(diamonds) > 200 {
		// 最多一次转移200枚
		return fmt.Errorf("diamonds quantity cannot over 200")
	}
	// check diamond name
	var realDiamonds = make([]fields.Bytes6, len(diamonds))
	for i, v := range diamonds {
		if !x16rs.IsDiamondValueString(v) {
			return fmt.Errorf("<%s> is not a diamond name", v)
		}
		realDiamonds[i] = fields.Bytes6(v)
	}

	var diamondOwnerAccount *account.Account = nil
	// address
	diamondOwnerStr := strings.TrimPrefix(CheckParamString(r, "diamond_owner_prikey", ""), "0x")
	diamondOwnerStr = strings.TrimPrefix(diamondOwnerStr, "0X")
	if len(diamondOwnerStr) > 0 {
		if len(diamondOwnerStr) != 64 {
			return fmt.Errorf("param main_prikey length error")
		}
		var prikeybts []byte
		if prikeybts, err = hex.DecodeString(diamondOwnerStr); err != nil {
			return fmt.Errorf("param main_prikey format error")
		}
		diamondOwnerAccount, err = account.GetAccountByPriviteKey(prikeybts)
		if err != nil {
			return err
		}
	}
	// 如果不传则为主地址
	if diamondOwnerAccount == nil {
		diamondOwnerAccount = mainAccount
	} else {
		// 加入签名私钥
		allprikey[string(diamondOwnerAccount.Address)] = diamondOwnerAccount.PrivateKey
	}
	if len(allprikey) > 1 {
		isMultiTrs = true // 主地址与钻石地址不一样时，为批量转移钻石
	}

	// to address
	addrStr := CheckParamString(r, "to_address", "")
	if len(addrStr) == 0 {
		return fmt.Errorf("param to_address must give")
	}
	to_addr, e8 := fields.CheckReadableAddress(addrStr)
	if e8 != nil {
		return e8
	}

	var actObj interfaces.Action = nil
	if isMultiTrs {
		// 批量转账
		actObj = &actions.Action_6_OutfeeQuantityDiamondTransfer{
			FromAddress:  diamondOwnerAccount.Address,
			ToAddress:    *to_addr,
			DiamondCount: fields.VarUint1(len(realDiamonds)),
			Diamonds:     realDiamonds,
		}
		//fmt.Println(diamondOwnerAccount.AddressReadable, to_addr.ToReadable(), len(realDiamonds), realDiamonds)
		//fmt.Println( `
		//// 批量转账
		//actObj = &actions.Action_6_OutfeeQuantityDiamondTransfer{
		//	FromAddress: diamondOwnerAccount.Address,
		//	ToAddress: *to_addr,
		//	DiamondCount: fields.VarUint1(len(realDiamonds)),
		//	Diamonds: realDiamonds,
		//}
		//` )
	} else {
		// 单个转账
		actObj = &actions.Action_5_DiamondTransfer{
			Diamond: fields.Bytes6(diamonds[0]),
			Address: *to_addr,
		}
		//fmt.Println( `
		//// 单个转账
		//actObj = &actions.Action_5_DiamondTransfer{
		//	Diamond: fields.Bytes6(diamonds[0]),
		//	Address: *to_addr,
		//}
		//` )
	}

	// 添加 act
	tx.AppendAction(actObj)

	// success
	return nil

}
