// Copyright 2012 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

package main

import (
	"github.com/evalgo/evapi"
	"github.com/evalgo/evapplication"
	"github.com/evalgo/evmail"
	"github.com/evalgo/evmonitor"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

func main() {
	mail := evmail.NewEVMailEmail()
	gobObjects := evapplication.NewEVApplicationGobRegisteredObjects()
	gobObjects.Append(mail)
	gobObjects.Append(evmail.NewEVMailImapEmail())
	gobObjects.Append(evmail.NewEVMailImapMailBoxes())
	gobObjects.Append(evmail.NewEVMailImapEmailHeader())
	gobObjects.Append(evmail.NewEVMailImapMailBoxInformation())
	gobObjects.Append(evmail.NewEVMailImapMailBoxAllMessages())
	gobObjects.RegisterAll()
	rpc.Register(mail)
	rpc.HandleHTTP()
	var ip string = ""
	ip, err := evapi.EVApiHostIp()
	if err != nil {
		log.Println("warning:", err)
		ip = "127.0.0.1"
	}
	service := evmonitor.NewEVMonitorService()
	service.Ip = ip
	hostname, err := evapi.EVApiHostName()
	if err != nil {
		log.Println("warning:", err)
	}
	service.Name = hostname
	service.Port = evapi.EVApiPortEVeMail
	service.Type = "evemail-rpc"
	log.Println("register evemail-rpc to monitoring...")
	_, err = evmonitor.EVMonitorRegisterService(service.Ip, service.Name, service.Ip, evapi.EVApiPortRedis, service)
	if err != nil {
		log.Println("warning:", err)
	}
	log.Println("starting evemail-rpc on  " + ip + ":" + strconv.Itoa(evapi.EVApiPortEVeMail) + "...")
	l, e := net.Listen("tcp", ip+":"+strconv.Itoa(evapi.EVApiPortEVeMail))
	if e != nil {
		log.Println("delete evemail-rpc from monitoring...")
		_, err := evmonitor.EVMonitorDeleteService(service.Ip, service.Name, service.Ip, evapi.EVApiPortRedis, service)
		if err != nil {
			log.Println("warning:", err)
		}

		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)
}
