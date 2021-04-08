package main

import (
	"fmt"
	"github.com/hacash/core/fields"
	"log"
	"net/http"
	"strconv"
)

func (api *Ranking) startHttpApiService() {

	port := api.http_api_listen_port
	if port == 0 {
		// 不启动服务器
		fmt.Println("config http_api_listen_port==0 do not start http api service.")
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ResponseData(w, ResponseCreateData("service", "hacash ranking service"))
	})

	// 路由
	mux.HandleFunc("/query", api.apiHandleFunc)

	// 设置监听的端口
	portstr := strconv.Itoa(port)
	server := &http.Server{
		Addr:    ":" + portstr,
		Handler: mux,
	}

	fmt.Println("[Hacash Ranking Service] Http api listen on port: " + portstr)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}

func (api *Ranking) apiHandleFunc(w http.ResponseWriter, r *http.Request) {

	api.dataChangeLocker.Lock()
	defer api.dataChangeLocker.Unlock()

	action := CheckParamString(r, "action", "")

	if action == "ranking" {

		kind := CheckParamString(r, "kind", "")
		if kind == "hacash" {

			list := make([]string, len(api.hacash_balance_ranking_100))
			for i, v := range api.hacash_balance_ranking_100 {
				percent := 0.0
				if api.current_circulation > 0 {
					percent = v.BalanceFloat64() / api.current_circulation * 100
				}
				list[i] = fmt.Sprintf("%s %.4f %.2f", v.Address.ToReadable(), v.BalanceFloat64(), percent)
			}
			ResponseList(w, list)

		} else if kind == "diamond" {

			list := make([]string, len(api.diamond_balance_ranking_100))
			for i, v := range api.diamond_balance_ranking_100 {
				percent := 0.0
				if api.minted_diamond > 0 {
					percent = float64(v.BalanceUint64()) / float64(api.minted_diamond) * 100
				}
				list[i] = fmt.Sprintf("%s %d %.2f", v.Address.ToReadable(), v.BalanceUint64(), percent)
			}
			ResponseList(w, list)

		} else if kind == "bitcoin" {

			list := make([]string, len(api.satoshi_balance_ranking_100))
			for i, v := range api.satoshi_balance_ranking_100 {
				percent := 0.0
				if api.minted_diamond > 0 {
					percent = float64(v.BalanceUint64()) / float64(api.transferred_bitcoin*100000000) * 100
				}
				list[i] = fmt.Sprintf("%s %.4f %.2f", v.Address.ToReadable(), float64(v.BalanceUint64())/100000000, percent)
			}
			ResponseList(w, list)

		} else {
			ResponseError(w, fmt.Errorf("cannot find kind <%s>", kind))
		}

	} else if action == "account_diamonds" {

		addrstr := CheckParamString(r, "address", "")
		_, e1 := fields.CheckReadableAddress(addrstr)
		if e1 != nil {
			ResponseErrorString(w, "address format error")
			return
		}
		// 查询账户的钻石列表
		diatable, hav1 := api.cache_update_diamonds[addrstr]
		if !hav1 {
			// 从 磁盘加载
			v, e := api.ldb.Get([]byte("ds"+addrstr), nil)
			//fmt.Println(string(v), e)
			if e == nil {
				diatable = v // load ok
			}
		}
		// 列出所有钻石
		data := ResponseCreateData("diamonds", string(diatable))
		ResponseData(w, data) // ok

	} else {
		ResponseError(w, fmt.Errorf("cannot find action <%s>", action))
	}

}
