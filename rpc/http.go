package rpc

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func (api *RpcService) RunHttpRpcService(port int) {

	api.initRoutes()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ResponseData(w, ResponseCreateData("service", "hacash node rpc"))
	}) // 查询
	mux.HandleFunc("/query", api.dealQuery) // 查询
	//mux.HandleFunc("/operate", api.dealOperate) // 写入
	//mux.HandleFunc("/operate_hex", api.dealOperateHex) // 写入

	portstr := strconv.Itoa(port)

	// 设置监听的端口
	server := &http.Server{
		Addr:    ":" + portstr,
		Handler: mux,
	}

	fmt.Println("[RPC Service] Http listen on port: " + portstr)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
			os.Exit(0)
		}
	}()
}
