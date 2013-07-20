// Copyright 2012 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

package evmail

import (
	"github.com/evalgo/everror"
	"github.com/evalgo/evlog"
	"github.com/evalgo/evmessage"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

var EVMailVersion string = "v1"

func (mail *EVMailEmail) createSendMail(r *http.Request) (*evmessage.EVMessage, *evmessage.EVMessage, string, error) {
	requestMsg, responseMsg := evmessage.EVMessageRpcClientInitialize("evmail")
	request, err := requestMsg.Body("requests").(*evmessage.EVMessageRequests).ById("evmail")
	if err != nil {
		responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(everror.NewFromError(err))
		return nil, responseMsg, "", everror.NewFromError(err)
	}
	if r.FormValue("request_id") == "" {
		err = everror.New("request_id is empty for service EVMail createSendMail")
		responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(everror.NewFromError(err))
		return nil, responseMsg, "", everror.NewFromError(err)
	}
	request.Id = r.FormValue("request_id")
	kvs := evmessage.NewEVMessageKeyValues()

	kvs.Append("user", r.FormValue("user"))
	kvs.Append("password", r.FormValue("password"))
	kvs.Append("server", r.FormValue("server"))
	kvs.Append("port", r.FormValue("port"))
	kvs.Append("to", r.FormValue("to"))
	kvs.Append("from", r.FormValue("from"))
	kvs.Append("subject", r.FormValue("subject"))
	kvs.Append("message", r.FormValue("message"))

	files := evmessage.NewEVMessageFiles()
	for key, _ := range r.MultipartForm.File {
		evlog.Println(key)
		f := evmessage.NewEVMessageFile()
		File, FileHeader, _ := r.FormFile(key)
		f.Name = FileHeader.Filename
		fContent, _ := ioutil.ReadAll(File)
		f.Content = fContent
		f.EncodeBase64()
		files.Append(f)
	}

	request.AppendToBody(files)
	request.AppendToBody(kvs)
	requestMsg.Body("requests").(*evmessage.EVMessageRequests).InProgress = request
	return requestMsg, responseMsg, "EVMailEmail.RpcSend", nil
}

func (mail *EVMailEmail) EVMessageHttpCreateRpcMessage(w http.ResponseWriter, r *http.Request) (*evmessage.EVMessage, *evmessage.EVMessage, string, error) {
	requestUrlPath := strings.TrimRight(r.URL.Path, "/")
	extension := filepath.Ext(requestUrlPath)
	requestUrlPath = strings.Replace(requestUrlPath, extension, "", 1)
	switch requestUrlPath {
	case "/" + EVMailVersion + "/emails", "/emails":
		switch r.Method {
		case "POST":
			return mail.createSendMail(r)
		default:
			return nil, nil, "", everror.New("the given request method+" + r.Method + " is not supported")
		}
	default:
		return nil, nil, "", everror.New("...")
	}
}

func (mail *EVMailEmail) EVMessageHttpRpcHandleResponse(w http.ResponseWriter, r *http.Request, responseMsg *evmessage.EVMessage) ([]byte, error) {
	requestUrlPath := strings.TrimRight(r.URL.Path, "/")
	extension := filepath.Ext(requestUrlPath)
	switch extension {
	case ".xml":
		evlog.Println("response format is XML")
		return responseMsg.ToXml()
	case ".json":
		evlog.Println("response format is JSON")
		return responseMsg.ToJson()
	}
	evlog.Println("response format is XML (default)")
	return responseMsg.ToXml()
}
