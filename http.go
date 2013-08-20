// Copyright 2012 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

// TODO finish HttpCreateRequest implementation

package evmail

import (
	"github.com/evalgo/everror"
	"github.com/evalgo/evmessage"
	"path/filepath"
	"strings"
)

var Version string = "v1"

type HttpRpcProxy struct{}

func NewHttpRpcProxy() *HttpRpcProxy {
	return new(HttpRpcProxy)
}

func (rpcP *HttpRpcProxy) HttpCreateRequest(vars map[string]string, req *evmessage.Request) (*evmessage.Request, error) {
	r := req.Body("http").(*evmessage.Http)
	extension := filepath.Ext(vars["action"])
	requestUrlPath := strings.Replace(strings.Join([]string{vars["version"], vars["service"], vars["action"]}, "/"), extension, "", 1)
	switch requestUrlPath {
	case Version + "/evemail/emails", "/evemail/emails":
		switch r.Method {
		case "POST":
			keyValuesI := evmessage.NewKeyValues()
			keyValuesI.ObjName = "rpc-request"
			keyValuesI.Append("RPC_ServiceName", "evemail-rpc")
			keyValuesI.Append("RPC_FuncName", "Email.RpcSend")
			req.AppendToBody(keyValuesI)
			return req, nil
		default:
			return nil, everror.New("the given request method+" + r.Method + " is not supported")
		}
	default:
		return nil, everror.New("URL path <" + r.URI + "> is not supported!")
	}
	return req, nil
}
