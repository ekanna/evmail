// Copyright 2012 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

package main

import (
	"github.com/evalgo/evapi"
	"github.com/evalgo/evapplication"
	"github.com/evalgo/everror"
	"github.com/evalgo/evlog"
	"github.com/evalgo/evmail"
	"github.com/evalgo/evmonitor"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

func main() {
	mail := evmail.NewEmail()
	gobObjects := evapplication.NewGobRegisteredObjects()
	gobObjects.Append(mail)
	gobObjects.Append(evmail.NewImapEmail())
	gobObjects.Append(evmail.NewImapMailBoxes())
	gobObjects.Append(evmail.NewImapEmailHeader())
	gobObjects.Append(evmail.NewImapMailBoxInformation())
	gobObjects.Append(evmail.NewImapMailBoxAllMessages())
	gobObjects.RegisterAll()
	rpc.Register(mail)
	rpc.HandleHTTP()
	var ip string = ""
	ip, err := evapi.HostIp()
	if err != nil {
		evlog.Println("warning:", everror.NewFromError(err))
		ip = "127.0.0.1"
	}
	service := evmonitor.NewService()
	service.Ip = ip
	hostname, err := evapi.HostName()
	if err != nil {
		evlog.Println("warning:", everror.NewFromError(err))
	}
	service.Name = hostname
	service.Port = evapi.PortEVeMail
	service.Type = "evemail-rpc"
	evlog.Println("register evemail-rpc to monitoring...")
	_, err = evmonitor.RegisterService(service.Ip, service.Name, service.Ip, evapi.PortRedis, service)
	if err != nil {
		evlog.Println("warning:", everror.NewFromError(err))
	}
	evlog.Println("starting evemail-rpc on  " + ip + ":" + strconv.Itoa(evapi.PortEVeMail) + "...")
	l, e := net.Listen("tcp", ip+":"+strconv.Itoa(evapi.PortEVeMail))
	if e != nil {
		evlog.Println("delete evemail-rpc from monitoring...")
		_, err := evmonitor.DeleteService(service.Ip, service.Name, service.Ip, evapi.PortRedis, service)
		if err != nil {
			evlog.Println("warning:", everror.NewFromError(err))
		}

		evlog.Fatal("listen error:", everror.NewFromError(e))
	}
	http.Serve(l, nil)
}
