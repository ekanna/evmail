// Copyright 2012 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

package evmail

import (
	"errors"
	"github.com/evalgo/evmessage"
	"log"
	"strconv"
	"strings"
)

func (mail *EVMailEmail) RpcSend(requestMsg *evmessage.EVMessage, responseMsg *evmessage.EVMessage) error {
	*responseMsg = *requestMsg
	responseMsg, err := evmessage.EVMessageRpcServiceInitialize(requestMsg, responseMsg)
	if err != nil {
		responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(err)
		return err
	}
	request := requestMsg.Body("requests").(*evmessage.EVMessageRequests).InProgress
	respObj := evmessage.NewEVMessageResponse()
	if request == nil {
		err = errors.New("there is no Requests.InProgress request set!")
		responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(err)
		return err
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
		responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(err)
		return err
	}
	email.Port = port
	email.To = strings.Split(kValues.ByKey("to"), " ")
	email.From = kValues.ByKey("from")
	email.Subject = kValues.ByKey("subject")
	email.Body = kValues.ByKey("message")
	files := request.Body("files").(*evmessage.EVMessageFiles)
	if files != nil {
		for _, f := range files.Files {
			log.Println("sending attachment:", f.Name)
			f.DecodeBase64()
			err := f.WriteFile("/tmp")
			if err != nil {
				responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(err)
				return err
			}
			err = email.Attach("/tmp/" + f.Name)
			if err != nil {
				responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(err)
				return err
			}
		}
	}
	err = email.Send()
	if err != nil {
		responseMsg.Body("errors").(*evmessage.EVMessageErrors).Append(err)
		return err
	}
	responseMsg.Body("responses").(*evmessage.EVMessageResponses).Append(respObj)
	return nil
}
