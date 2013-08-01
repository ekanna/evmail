// Copyright 2012 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

package evmail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/evalgo/everror"
	"io/ioutil"
	"net/smtp"
	"path/filepath"
	"time"
)

func (Mail *Email) Send() error {
	var recipients = ""
	boundary := "f46d043c813270fc6b04c2d223da"
	for _, i := range Mail.To {
		recipients += i + ","
	}
	content := bytes.NewBuffer(nil)
	content.WriteString("Date: " + time.Now().Format("Mon, 2 Jan 2006 15:04:05 -0700") + "\r\n")
	content.WriteString("From: " + Mail.From + "\r\n")
	content.WriteString("To: " + recipients + "\r\n")
	//content.WriteString("Reply-To: "+e.From+"\r\n")
	content.WriteString("Subject: " + Mail.Subject + "\r\n")
	content.WriteString("MIME-Version: 1.0\r\n")
	if len(Mail.Attachments) > 0 {
		content.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n")
		content.WriteString("--" + boundary + "\r\n")
	}
	content.WriteString("Content-Type: text/plain; charset=utf-8\n")
	content.WriteString(Mail.Body)
	if len(Mail.Attachments) > 0 {
		for k, v := range Mail.Attachments {
			content.WriteString("\r\n\r\n--" + boundary + "\r\n")
			content.WriteString("Content-Type: application/octet-stream\r\n")
			content.WriteString("Content-Transfer-Encoding: base64\r\n")
			content.WriteString("Content-Disposition: attachment; filename=\"" + k + "\"\r\n\r\n")

			buffer := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
			base64.StdEncoding.Encode(buffer, v)
			content.Write(buffer)
			content.WriteString("\r\n--" + boundary)
		}

		content.WriteString("--")
	}
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", Mail.Server, Mail.Port),
		SuperPlainAuth(Mail.User, Mail.Password),
		Mail.From,
		Mail.To,
		content.Bytes())
	if err != nil {
		return everror.NewFromError(err)
	}
	return nil
}

func (Mail *Email) Attach(file string) error {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return everror.NewFromError(err)
	}

	_, fileName := filepath.Split(file)
	Mail.Attachments[fileName] = buffer
	return nil
}

func SuperPlainAuth(UserName string, Password string) smtp.Auth {
	return &superPlainAuth{UserName, Password}
}

func (Auth *superPlainAuth) Start(Server *smtp.ServerInfo) (string, []byte, error) {
	Resp := []byte("" + "\x00" + Auth.UserName + "\x00" + Auth.Password)
	return "PLAIN", Resp, nil
}

func (Auth *superPlainAuth) Next(FromServer []byte, More bool) ([]byte, error) {
	if More {
		return nil, everror.New("unexpected server challenge")
	}
	return nil, nil
}
