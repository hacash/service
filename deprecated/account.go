package rpc

import (
	"encoding/hex"
	"github.com/hacash/core/account"
)

func newAccountByPassword(params map[string]string) map[string]string {
	result := make(map[string]string)

	passstr, ok1 := params["password"]
	if !ok1 {
		result["err"] = "password must"
		return result
	}

	// Create account
	acc := account.CreateAccountByPassword(passstr)

	result["address"] = string(acc.AddressReadable)
	result["public_key"] = hex.EncodeToString(acc.PublicKey)
	result["private_key"] = hex.EncodeToString(acc.PrivateKey)

	return result
}

// Random creation
func newAccount(params map[string]string) map[string]string {
	result := make(map[string]string)
	// Create account
	acc := account.CreateNewRandomAccount()

	result["address"] = string(acc.AddressReadable)
	result["public_key"] = hex.EncodeToString(acc.PublicKey)
	result["private_key"] = hex.EncodeToString(acc.PrivateKey)

	return result
}
