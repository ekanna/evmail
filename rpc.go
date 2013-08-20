// Copyright 2012 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

package evmail

import (
	"github.com/evalgo/everror"
	"github.com/evalgo/evlog"
	"github.com/evalgo/evmessage"
	"path/filepath"
	"strconv"
	"strings"
)

func (mail *Email) RpcSend(requestMsg *evmessage.Message, responseMsg *evmessage.Message) (e error) {
	evlog.Println("running evemail-rpc.RpcSend ...")
	defer everror.ResetAllErrors()
	defer func() {
		// handle panic
		if r := recover(); r != nil {
			everror.NewFromString(r.(error).Error(), everror.FATAL)
			e = nil
		}
		// write all errors to the response message
		responseMsg = LoadMsgErrors(responseMsg)
		evlog.Println("ALL ERRORS:", len(everror.AllErrors.Errors))
	}()
	evlog.Println("reset all errors...")
	everror.ResetAllErrors()
	evlog.Println("starting send ...")
	evlog.Println(requestMsg.Body("requests").(*evmessage.Requests).InProgress.Body("http"))
	responseMsg, err := evmessage.RpcServiceInitialize(requestMsg, responseMsg)
	if err != nil {
		return everror.NewFromError(err, everror.ERROR)
	}
	request := requestMsg.Body("requests").(*evmessage.Requests).InProgress
	respObj := evmessage.NewResponse()
	if request == nil {
		return everror.New("there is no Requests.InProgress request set!")
	}
	respObj.Id = request.Id
	respObj.Order = request.Order
	kValues := requestMsg.Body("requests").(*evmessage.Requests).InProgress.Body("http").(*evmessage.Http)
	email := NewEmail()
	email.User = kValues.FormValue("user")
	email.Password = kValues.FormValue("password")
	email.Server = kValues.FormValue("server")
	port, err := strconv.Atoi(kValues.FormValue("port"))
	if err != nil {
		return everror.NewFromError(err, everror.ERROR)
	}
	email.Port = port
	email.To = strings.Split(kValues.FormValue("to"), " ")
	email.From = kValues.FormValue("from")
	email.Subject, err = kValues.FormValueBase64("subject")
	email.Body, err = kValues.FormValueBase64("message")

	for _, f := range kValues.Files {
		evlog.Println("sending attachment:", f.Name)
		f.DecodeBase64()
		err := f.WriteFile("/tmp")
		if err != nil {
			return everror.NewFromError(err, everror.ERROR)
		}
		err = email.Attach("/tmp/" + f.Name)
		if err != nil {
			return everror.NewFromError(err, everror.ERROR)
		}
	}

	err = email.Send()
	if err != nil {
		return everror.NewFromError(err, everror.ERROR)
	}
	mimeType := filepath.Ext(kValues.URI)
	keyValues := evmessage.NewKeyValues()
	keyValues.ObjName = "response"
	switch mimeType {
	case ".txt":
		keyValues.Append("response.mime.type", "text/plain")
		keyValues.Append("response.content", "email was sent successfully")
	case ".html":
		keyValues.Append("response.mime.type", "text/html")
		keyValues.Append("response.content", "email was sent successfully")
	case ".json":
		keyValues.Append("response.mime.type", "application/json")
		keyValues.AppendBase64("response.content", "{response:{message:\"email was sent successfully\"}}")
	default:
		keyValues.Append("response.mime.type", "text/xml")
		keyValues.Append("response.content", "<response><message>email was sent successfully</message></response>")

	}
	respObj.AppendToBody(keyValues)
	responseMsg.Body("responses").(*evmessage.Responses).Append(respObj)
	return nil
}
