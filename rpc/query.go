package rpc

import (
	"io/ioutil"
	"net/http"
)

func (api *RpcService) dealRoutes(routes map[string]func(*http.Request, http.ResponseWriter, []byte), w http.ResponseWriter, r *http.Request, gotbodybytes bool) {
	var err error
	var bodybytes []byte = nil

	if gotbodybytes {
		bodybytes, err = ioutil.ReadAll(r.Body)
		if err != nil {
			ResponseError(w, err)
			return
		}
	}

	r.ParseForm()

	actionName := r.FormValue("action")

	if len(actionName) == 0 {
		ResponseErrorString(w, "param 'action' must give.")
		return
	}

	action, actok := routes[actionName]
	if !actok {
		ResponseErrorString(w, "not find action <"+actionName+">.")
		return
	}

	// call action
	action(r, w, bodybytes)
}

func (api *RpcService) dealQuery(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.queryRoutes, w, r, false)
}

func (api *RpcService) dealCreate(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.createRoutes, w, r, false)
}

func (api *RpcService) dealSubmit(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.submitRoutes, w, r, true)
}

func (api *RpcService) dealOperate(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.operateRoutes, w, r, false)
}
