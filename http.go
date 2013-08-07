// Copyright 2012 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

// TODO finish HttpCreateRequest implementation

package evmail

import (
	"github.com/evalgo/everror"
	"github.com/evalgo/evlog"
	"github.com/evalgo/evmessage"
	"io/ioutil"
	"net/http"
	//"path/filepath"
	//"strings"
)

var Version string = "v1"

type HttpRpcProxy struct{}

func NewHttpRpcProxy() *HttpRpcProxy {
	return new(HttpRpcProxy)
}

func (rpcP *HttpRpcProxy) createSendMail(r *http.Request) (*evmessage.Message, *evmessage.Message, string, error) {
	requestMsg, responseMsg := evmessage.RpcClientInitialize("evmail")
	request, err := requestMsg.Body("requests").(*evmessage.Requests).ById("evmail")
	if err != nil {
		responseMsg.Body("errors").(*evmessage.Errors).Append(everror.NewFromError(err))
		return nil, responseMsg, "", everror.NewFromError(err)
	}
	if r.FormValue("request_id") == "" {
		err = everror.New("request_id is empty for service  createSendMail")
		responseMsg.Body("errors").(*evmessage.Errors).Append(everror.NewFromError(err))
		return nil, responseMsg, "", everror.NewFromError(err)
	}
	request.Id = r.FormValue("request_id")
	kvs := evmessage.NewKeyValues()

	kvs.Append("user", r.FormValue("user"))
	kvs.Append("password", r.FormValue("password"))
	kvs.Append("server", r.FormValue("server"))
	kvs.Append("port", r.FormValue("port"))
	kvs.Append("to", r.FormValue("to"))
	kvs.Append("from", r.FormValue("from"))
	kvs.Append("subject", r.FormValue("subject"))
	kvs.Append("message", r.FormValue("message"))

	files := evmessage.NewFiles()
	for key, _ := range r.MultipartForm.File {
		evlog.Println(key)
		f := evmessage.NewFile()
		File, FileHeader, _ := r.FormFile(key)
		f.Name = FileHeader.Filename
		fContent, _ := ioutil.ReadAll(File)
		f.Content = fContent
		f.EncodeBase64()
		files.Append(f)
	}

	request.AppendToBody(files)
	request.AppendToBody(kvs)
	requestMsg.Body("requests").(*evmessage.Requests).InProgress = request
	return requestMsg, responseMsg, "Email.RpcSend", nil
}

func (rpcP *HttpRpcProxy) HttpCreateRequest(vars map[string]string, r *evmessage.Request) (*evmessage.Request, error) {
	// requestUrlPath := strings.TrimRight(r.URL.Path, "/")
	// extension := filepath.Ext(requestUrlPath)
	// requestUrlPath = strings.Replace(requestUrlPath, extension, "", 1)
	// switch requestUrlPath {
	// case "/" + Version + "/evemail/emails", "/evemail/emails":
	// 	switch r.Method {
	// 	case "POST":
	// 		return mail.createSendMail(r)
	// 	default:
	// 		return nil, nil, "", everror.New("the given request method+" + r.Method + " is not supported")
	// 	}
	// default:
	// 	return nil, nil, "", everror.New("URL path <" + r.URL.Path + "> is not supported!")
	// }
	return r, nil
}
