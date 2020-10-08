package rpc

import "net/http"

func (api *RpcService) dealQuery(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	actionName := r.FormValue("action")

	if len(actionName) == 0 {
		ResponseErrorString(w, "param 'action' must give.")
		return
	}

	action, actok := api.queryRoutes[actionName]
	if !actok {
		ResponseErrorString(w, "not find antion <"+actionName+">.")
		return
	}

	// call action
	action(r, w)
}
