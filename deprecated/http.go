package rpc

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
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

	api.routeOperateRequest(request, response, binary.BigEndian.Uint32(bodybytes[0:4]), bodybytes[4:])
}

func (api *DeprecatedApiService) dealOperate(response http.ResponseWriter, request *http.Request) {
	bodybytes, e1 := ioutil.ReadAll(request.Body)
	//fmt.Println(bodybytes, e1)
	if e1 != nil {
		response.Write([]byte("body error"))
		return
	}

	if len(bodybytes) < 4 {
		response.Write([]byte("body length less than 4"))
		return
	}

	api.routeOperateRequest(request, response, binary.BigEndian.Uint32(bodybytes[0:4]), bodybytes[4:])
}

func parseRequestQuery(request *http.Request) map[string]string {
	request.ParseForm()
	params := make(map[string]string, 0)
	for k, v := range request.Form {
		params[k] = strings.Join(v, "")
	}

	return params
}

func (api *DeprecatedApiService) RunHttpRpcService(port int) {

	api.initRoutes()

	mux := http.NewServeMux()

	mux.Handle("/websocket", websocket.Handler(api.webSocketHandler))

	mux.HandleFunc("/", api.dealHome)
	mux.HandleFunc("/query", api.dealQuery)
	mux.HandleFunc("/operate", api.dealOperate)
	mux.HandleFunc("/operatehex", api.dealOperateHex)
	portstr := strconv.Itoa(port)

	// Set listening port
	server := &http.Server{
		Addr:    ":" + portstr,
		Handler: mux,
	}

	fmt.Println("[Deprecated Api Service] Http listen on port: " + portstr)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe Error: ", err)
		} else {
			fmt.Println("RunHttpRpcService on " + portstr)
		}
	}()
}
