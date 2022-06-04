package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/account"
	"github.com/hacash/core/actions"
	"github.com/hacash/core/fields"
	"github.com/hacash/core/interfacev2"
	"github.com/hacash/core/transactions"
	"net/http"
	"strings"
	"time"
)

// Create value transfer transaction
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
		feeAmt, err = fields.NewAmountFromMeiStringUnsafe(feeStr)
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

	newTrs.Timestamp = fields.BlockTxTimestamp(timestamp)
	newTrs.Fee = *feeAmt

	// create action
	transferKind := CheckParamString(r, "transfer_kind", "")
	var kinderr error
	switch transferKind {
	case "hacash":
		// Transfer HAC
		kinderr = appendActionSimpleTransferHacash(r, isUnitMei, allPrivateKeyBytes, mainAccount, newTrs)
	case "satoshi":
		// Transfer BTC
		kinderr = appendActionSimpleTransferSatoshi(r, isUnitMei, allPrivateKeyBytes, mainAccount, newTrs)
	case "diamond":
		// Transfer diamond
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
		amountAmt, err = fields.NewAmountFromMeiStringUnsafe(amountStr)
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

	var actObj = &actions.Action_1_SimpleToTransfer{
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
		ToAddress: *to_addr,
		Amount:    fields.Satoshi(amountAmt),
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

	var diamonds = fields.NewEmptyDiamondListMaxLen200()
	e0 := diamonds.ParseHACDlistBySplitCommaFromString(diamondsStr)
	if e0 != nil {
		return e0
	}

	if diamonds.Count > 1 {
		isMultiTrs = true // More than one batch transfer
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

	// If not, it will be the primary address
	if diamondOwnerAccount == nil {
		diamondOwnerAccount = mainAccount
	} else {
		// Add signature private key
		allprikey[string(diamondOwnerAccount.Address)] = diamondOwnerAccount.PrivateKey
	}

	if len(allprikey) > 1 {
		isMultiTrs = true // When the main address is different from the diamond address, the diamonds are transferred in batches
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

	var actObj interfacev2.Action = nil
	if isMultiTrs {
		// Batch transfer
		actObj = &actions.Action_6_OutfeeQuantityDiamondTransfer{
			FromAddress: diamondOwnerAccount.Address,
			ToAddress:   *to_addr,
			DiamondList: *diamonds,
		}
	} else {
		// Single transfer
		actObj = &actions.Action_5_DiamondTransfer{
			Diamond:   fields.DiamondName(diamonds.Diamonds[0]),
			ToAddress: *to_addr,
		}
	}

	// Add act
	tx.AppendAction(actObj)

	// success
	return nil
}
