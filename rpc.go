// Copyright 2012 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

package evmail

import (
	"github.com/evalgo/everror"
	"github.com/evalgo/evlog"
	"github.com/evalgo/evmessage"
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
	responseMsg, err := evmessage.RpcServiceInitialize(requestMsg, responseMsg)
	if err != nil {
		responseMsg.Body("errors").(*evmessage.Errors).Append(everror.NewFromError(err, everror.ERROR))
		return everror.NewFromError(err, everror.ERROR)
	}
	request := requestMsg.Body("requests").(*evmessage.Requests).InProgress
	respObj := evmessage.NewResponse()
	if request == nil {
		err = everror.New("there is no Requests.InProgress request set!")
		responseMsg.Body("errors").(*evmessage.Errors).Append(everror.NewFromError(err, everror.ERROR))
		return everror.NewFromError(err, everror.ERROR)
	}
	respObj.Id = request.Id
	respObj.Order = request.Order
	kValues := request.Body("keyvalues").(*evmessage.KeyValues)
	email := NewEmail()
	email.User = kValues.ByKey("user")
	email.Password = kValues.ByKey("password")
	email.Server = kValues.ByKey("server")
	port, err := strconv.Atoi(kValues.ByKey("port"))
	if err != nil {
		responseMsg.Body("errors").(*evmessage.Errors).Append(everror.NewFromError(err, everror.ERROR))
		return everror.NewFromError(err, everror.ERROR)
	}
	email.Port = port
	email.To = strings.Split(kValues.ByKey("to"), " ")
	email.From = kValues.ByKey("from")
	email.Subject = kValues.ByKey("subject")
	email.Body = kValues.ByKey("message")
	files := request.Body("files").(*evmessage.Files)
	if files != nil {
		for _, f := range files.Files {
			evlog.Println("sending attachment:", f.Name)
			f.DecodeBase64()
			err := f.WriteFile("/tmp")
			if err != nil {
				responseMsg.Body("errors").(*evmessage.Errors).Append(everror.NewFromError(err, everror.ERROR))
				return everror.NewFromError(err, everror.ERROR)
			}
			err = email.Attach("/tmp/" + f.Name)
			if err != nil {
				responseMsg.Body("errors").(*evmessage.Errors).Append(everror.NewFromError(err, everror.ERROR))
				return everror.NewFromError(err, everror.ERROR)
			}
		}
	}
	err = email.Send()
	if err != nil {
		responseMsg.Body("errors").(*evmessage.Errors).Append(everror.NewFromError(err, everror.ERROR))
		return everror.NewFromError(err, everror.ERROR)
	}
	keyValues := evmessage.NewKeyValues()
	keyValues.Append("Message", "Email was sent successfully")
	respObj.AppendToBody(keyValues)
	responseMsg.Body("responses").(*evmessage.Responses).Append(respObj)
	return nil
}
