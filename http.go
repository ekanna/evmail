// Copyright 2012 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

package evmail

import (
	"errors"
	"github.com/evalgo/evmessage"
	"io/ioutil"
	"log"
	"net/http"
)

var EVMailVersion string = "v1"

func (mail *EVMailEmail) createSendMail(r *http.Request) (*evmessage.EVMessage, *evmessage.EVMessage, string, error) {
	requestMsg, responseMsg := evmessage.EVMessageRpcClientInitialize("evmail")
	request, err := requestMsg.Body("requests").(*evmessage.EVMessageRequests).ById("evmail")
	if err != nil {
		responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(err)
		return nil, responseMsg, "", err
	}
	if r.FormValue("request_id") == "" {
		err = errors.New("request_id is empty for service EVMail createSendMail")
		responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(err)
		return nil, responseMsg, "", err
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
		log.Println(key)
		f := evmessage.NewEVMessageFile()
		f.Name = key
		File, _, _ := r.FormFile(key)
		fContent, _ := ioutil.ReadAll(File)
		f.Content = fContent
		f.EncodeBase64()
		files.Append(f)
	}

	requestMsg.AppendToBody(files)
	request.AppendToBody(kvs)
	requestMsg.Body("requests").(*evmessage.EVMessageRequests).InProgress = request
	return requestMsg, responseMsg, "EVMailEmail.RpcSend", nil
}

func (mail *EVMailEmail) EVMessageHttpCreateRpcMessage(w http.ResponseWriter, r *http.Request) (*evmessage.EVMessage, *evmessage.EVMessage, string, error) {
	switch r.URL.Path {
	case "/" + EVMailVersion + "/emails", "/" + EVMailVersion + "/emails/", "/emails", "/emails/":
		switch r.Method {
		case "POST":
			return mail.createSendMail(r)
		default:
			return nil, nil, "", errors.New("the given request method+" + r.Method + " is not supported")
		}
	default:
		return nil, nil, "", errors.New("...")
	}
}

func (mail *EVMailEmail) EVMessageHttpRpcHandleResponse(w http.ResponseWriter, r *http.Request, responseMsg *evmessage.EVMessage) ([]byte, error) {
	return responseMsg.ToXml()
}
