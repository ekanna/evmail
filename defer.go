// Copyright 2012 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

package evmail

import (
	"github.com/evalgo/everror"
	"github.com/evalgo/evmessage"
)

func LoadMsgErrors(msg *evmessage.Message) *evmessage.Message {
	for _, err := range everror.AllErrors.Errors {
		msg.Body("errors").(*evmessage.Errors).Append(err)
	}
	return msg
}
