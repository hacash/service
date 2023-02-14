package rpc

import (
	"encoding/hex"
	"fmt"
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

////////////////////////////////

// total account
func (api *DeprecatedApiService) getTotalNonEmptyAccount(params map[string]string) map[string]string {
	result := make(map[string]string)
	//
	var tts = api.blockchain.GetChainEngineKernel().CurrentState().GetTotalNonEmptyAccountStatistics()

	result["total"] = fmt.Sprintf("%d", tts[0])
	result["hac"] = fmt.Sprintf("%d", tts[1])
	result["btc"] = fmt.Sprintf("%d", tts[2])
	result["hacd"] = fmt.Sprintf("%d", tts[3])

	return result
}
