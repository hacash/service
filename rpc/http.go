package rpc

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func (api *RpcService) RunHttpRpcService(port int) {
	api.initRoutes()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ResponseData(w, ResponseCreateData("service", "hacash node rpc"))
	})

	// route
	mux.HandleFunc("/query", api.dealQuery)     // query
	mux.HandleFunc("/create", api.dealCreate)   // establish
	mux.HandleFunc("/submit", api.dealSubmit)   // Submit
	mux.HandleFunc("/operate", api.dealOperate) // modify

	// Set listening port
	portstr := strconv.Itoa(port)
	server := &http.Server{
		Addr:    ":" + portstr,
		Handler: mux,
	}

	fmt.Println("[RPC Service] Http listen on port: " + portstr)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("ListenAndServe Error: ", err)
			return
		}
	}()
}
