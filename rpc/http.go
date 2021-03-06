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
	})

	// 路由
	mux.HandleFunc("/query", api.dealQuery)     // 查询
	mux.HandleFunc("/create", api.dealCreate)   // 创建
	mux.HandleFunc("/submit", api.dealSubmit)   // 提交
	mux.HandleFunc("/operate", api.dealOperate) // 修改

	// 设置监听的端口
	portstr := strconv.Itoa(port)
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
