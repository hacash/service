package rpc

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/hacash/node/websocket"

	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func (api *DeprecatedApiService) dealQuery(response http.ResponseWriter, request *http.Request) {
	params := parseRequestQuery(request)
	if _, ok := params["action"]; !ok {
		response.Write([]byte("must action"))
		return
	}

	// call controller
	routeQueryRequest(params["action"], params, response, request)

}

func (api *DeprecatedApiService) dealOperateHex(response http.ResponseWriter, request *http.Request) {

	bodyhexs, e1 := ioutil.ReadAll(request.Body)
	if e1 != nil {
		response.Write([]byte("body error"))
		return
	}
	bodybytes, e0 := hex.DecodeString(string(bodyhexs))
	if e0 != nil {
		response.Write([]byte("body hex format error"))
		return
	}
	if len(bodybytes) < 4 {
		response.Write([]byte("body length less than 4"))
		return
	}
	api.routeOperateRequest(response, binary.BigEndian.Uint32(bodybytes[0:4]), bodybytes[4:])
}

func (api *DeprecatedApiService) dealOperate(response http.ResponseWriter, request *http.Request) {

	bodybytes, e1 := ioutil.ReadAll(request.Body)
	if e1 != nil {
		response.Write([]byte("body error"))
		return
	}
	if len(bodybytes) < 4 {
		response.Write([]byte("body length less than 4"))
		return
	}
	api.routeOperateRequest(response, binary.BigEndian.Uint32(bodybytes[0:4]), bodybytes[4:])
}

func parseRequestQuery(request *http.Request) map[string]string {
	request.ParseForm()
	params := make(map[string]string, 0)
	for k, v := range request.Form {
		//fmt.Println("key:", k)
		//fmt.Println("val:", strings.Join(v, ""))
		params[k] = strings.Join(v, "")
	}
	return params
}

func (api *DeprecatedApiService) RunHttpRpcService(port int) {

	api.initRoutes()

	mux := http.NewServeMux()

	mux.Handle("/websocket", websocket.Handler(api.webSocketHandler))

	mux.HandleFunc("/", api.dealHome)                 //设置访问的路由
	mux.HandleFunc("/query", api.dealQuery)           //设置访问的路由
	mux.HandleFunc("/operate", api.dealOperate)       //设置访问的路由
	mux.HandleFunc("/operatehex", api.dealOperateHex) //设置访问的路由

	//http.HandleFunc("/minerpool", minerPoolStatisticsAutoTransfer)       //设置访问的路由
	//http.HandleFunc("/minerpool/transactions", minerPoolAllTransactions) //设置访问的路由
	//http.HandleFunc("/minerpool/statistics", minerPoolStatistics) //设置访问的路由

	portstr := strconv.Itoa(port)

	// 设置监听的端口
	server := &http.Server{
		Addr:    ":" + portstr,
		Handler: mux,
	}

	fmt.Println("[Deprecated Api Service] Http listen on port: " + portstr)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
			os.Exit(0)
		} else {
			fmt.Println("RunHttpRpcService on " + portstr)
		}
	}()
}
