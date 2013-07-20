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

func (mail *EVMailEmail) RpcSend(requestMsg *evmessage.EVMessage, responseMsg *evmessage.EVMessage) error {
	*responseMsg = *requestMsg
	responseMsg, err := evmessage.EVMessageRpcServiceInitialize(requestMsg, responseMsg)
	if err != nil {
		responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(everror.NewFromError(err))
		return everror.NewFromError(err)
	}
	request := requestMsg.Body("requests").(*evmessage.EVMessageRequests).InProgress
	respObj := evmessage.NewEVMessageResponse()
	if request == nil {
		err = everror.New("there is no Requests.InProgress request set!")
		responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(everror.NewFromError(err))
		return everror.NewFromError(err)
	}
	respObj.Id = request.Id
	respObj.Order = request.Order
	kValues := request.Body("keyvalues").(*evmessage.EVMessageKeyValues)
	email := NewEVMailEmail()
	email.User = kValues.ByKey("user")
	email.Password = kValues.ByKey("password")
	email.Server = kValues.ByKey("server")
	port, err := strconv.Atoi(kValues.ByKey("port"))
	if err != nil {
		responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(everror.NewFromError(err))
		return everror.NewFromError(err)
	}
	email.Port = port
	email.To = strings.Split(kValues.ByKey("to"), " ")
	email.From = kValues.ByKey("from")
	email.Subject = kValues.ByKey("subject")
	email.Body = kValues.ByKey("message")
	files := request.Body("files").(*evmessage.EVMessageFiles)
	if files != nil {
		for _, f := range files.Files {
			evlog.Println("sending attachment:", f.Name)
			f.DecodeBase64()
			err := f.WriteFile("/tmp")
			if err != nil {
				responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(everror.NewFromError(err))
				return everror.NewFromError(err)
			}
			err = email.Attach("/tmp/" + f.Name)
			if err != nil {
				responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(everror.NewFromError(err))
				return everror.NewFromError(err)
			}
		}
	}
	err = email.Send()
	if err != nil {
		responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(everror.NewFromError(err))
		return everror.NewFromError(err)
	}
	responseMsg.Body("responses").(*evmessage.EVMessageResponses).Append(respObj)
	return nil
}
