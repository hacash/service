package rpc

import "net/http"

func (api *RpcService) dealRoutes(routes map[string]func(*http.Request, http.ResponseWriter), w http.ResponseWriter, r *http.Request) {

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
	action(r, w)
}

func (api *RpcService) dealQuery(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.queryRoutes, w, r)
}

func (api *RpcService) dealCreate(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.createRoutes, w, r)
}

func (api *RpcService) dealSubmit(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.submitRoutes, w, r)
}

func (api *RpcService) dealOperate(w http.ResponseWriter, r *http.Request) {
	api.dealRoutes(api.operateRoutes, w, r)
}
