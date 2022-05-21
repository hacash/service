package rpc

import (
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/account"
)

func (api *DeprecatedApiService) readPasswordOrPriviteKeyParamBeAccount(params map[string]string, keyname string) (*account.Account, error) {

	password_or_privatekey, ok1 := params[keyname]
	if !ok1 {
		return nil, fmt.Errorf("param " + keyname + " must")
	}

	var acc *account.Account = nil
	privatekey, e2 := hex.DecodeString(password_or_privatekey)
	if len(password_or_privatekey) == 64 && e2 == nil {
		acc, e2 = account.GetAccountByPriviteKey(privatekey)
		if e2 != nil {
			return nil, fmt.Errorf("Privite Key Error")
		}
	} else {
		//fmt.Println(password_or_privatekey)
		acc = account.CreateAccountByPassword(password_or_privatekey)
		//fmt.Println(string(acc.AddressReadable))
		//fmt.Println(params["from"])
	}

	return acc, nil
}
