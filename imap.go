// Copyright 2012 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

package evmail

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type EVMailEmail struct {
	Server      string   `xml:"Server"`
	Port        int      `xml:"Port"`
	User        string   `xml:"User"`
	Password    string   `xml:"Password"`
	From        string   `xml:"From"`
	To          []string `xml:"To>Recipient"`
	Subject     string   `xml:"Subject"`
	Body        string   `xml:"Text"`
	Html        string   `xml:"Html"`
	Raw         string   `xml:"Raw"`
	Attachments map[string][]byte
}

func NewEVMailEmail() *EVMailEmail {
	email := new(EVMailEmail)
	email.Attachments = make(map[string][]byte, 0)
	return email
}

type EVMailImapEmail struct {
	Id   int          `xml:"Id"`
	Body *EVMailEmail `xml:"Body"`
}

func NewEVMailImapEmail() *EVMailImapEmail {
	email := new(EVMailImapEmail)
	email.Id = 0
	email.Body = NewEVMailEmail()
	return email
}

type superPlainAuth struct {
	UserName string `xml:"UserName"`
	Password string `xml:"Password"`
}

type EVMailImapMailBoxInformation struct {
	Flags       []string `xml:"Flags>Flag"`
	Mails       int      `xml:"Mails"`
	Recent      int      `xml:"Recent"`
	NonExistent bool     `xml:"NonExistent"`
	Path        string   `xml:"Path"`
}

func NewEVMailImapMailBoxInformation() *EVMailImapMailBoxInformation {
	info := new(EVMailImapMailBoxInformation)
	info.Flags = make([]string, 0)
	info.Mails = 0
	info.Recent = 0
	info.NonExistent = false
	info.Path = ""
	return info
}

type EVMailImapMailBoxes struct {
	MailBoxes []EVMailImapMailBoxInformation `xml:"Mailboxes>Mailbox"`
}

func NewEVMailImapMailBoxes() *EVMailImapMailBoxes {
	mBoxes := new(EVMailImapMailBoxes)
	mBoxes.MailBoxes = make([]EVMailImapMailBoxInformation, 0)
	return mBoxes
}

type EVMailImapEmailHeader struct {
	Id                    int    `xml:"Id"`
	DeliverdTo            string `xml:"DeliverdTo"`
	Received              string `xml:"Received"`
	ReturnPath            string `xml:"ReturnPath"`
	ReceivedSPF           string `xml:"Received-SPF"`
	AuthenticationResults string `xml:"Authentication-Results"`
	MimeVersion           string `xml:"Mime-Version"`
	Date                  string `xml:"Date"`
	MessageID             string `xml:"Message-ID"`
	Subject               string `xml:"Subject"`
	From                  string `xml:"From"`
	To                    string `xml:"To"`
	ContentType           string `xml:"Content-Type"`
	XGmMessageState       string `xml:"X-Gm-Message-State"`
}

func NewEVMailImapEmailHeader() *EVMailImapEmailHeader {
	emailH := new(EVMailImapEmailHeader)
	emailH.Id = 0
	emailH.DeliverdTo = ""
	emailH.Received = ""
	emailH.ReturnPath = ""
	emailH.ReceivedSPF = ""
	emailH.AuthenticationResults = ""
	emailH.MimeVersion = ""
	emailH.Date = ""
	emailH.MessageID = ""
	emailH.Subject = ""
	emailH.From = ""
	emailH.To = ""
	emailH.ContentType = ""
	emailH.XGmMessageState = ""
	return emailH
}

type EVMailImapMailBoxAllMessages struct {
	Headers []EVMailImapEmailHeader `xml:"Headers>Header"`
}

func NewEVMailImapMailBoxAllMessages() *EVMailImapMailBoxAllMessages {
	allM := new(EVMailImapMailBoxAllMessages)
	allM.Headers = make([]EVMailImapEmailHeader, 0)
	return allM
}

func ImapSendServerMessage(conn *tls.Conn, message string) (string, error) {
	n, err := io.WriteString(conn, message)
	if err != nil {
		log.Fatalf("client: write:%v::%s", n, err)
	}
	return "", nil
}

func ImapReadServerMessage(conn *tls.Conn) (string, int) {
	replyMessage := make([]byte, 4096)
	sn, _ := conn.Read(replyMessage)
	return string(replyMessage[:sn]), sn
}

func ImapReadMailBoxMessagesHeader(conn *tls.Conn) (string, error) {
	replyHeader := make([]byte, 4096)
	replyString := ""
	for {
		sn, _ := conn.Read(replyHeader)
		replyMessage := string(replyHeader[:sn])
		replyString = strings.Join([]string{replyString, replyMessage}, "")
		if true == strings.Contains(replyString, strings.Join([]string{".", "OK", "Success"}, " ")) {
			return replyString, nil
		}
	}
	return replyString, nil
}

func ImapGetAllMailboxes(conn *tls.Conn, Path string) []string {
	AllMailBoxes := make([]string, 0)
	PathOnTheServer := strings.Join([]string{"\"", Path, "\""}, "")
	ServerRequestMessage := strings.Join([]string{".", "LIST", PathOnTheServer, "\"*\"\r\n"}, " ")
	ImapSendServerMessage(conn, ServerRequestMessage)
	AllMailBoxesString, _ := ImapReadServerMessage(conn)
	log.Print(AllMailBoxesString)
	AllMailBoxesRaw := strings.Split(AllMailBoxesString, "\r\n")
	for _, MailBoxRaw := range AllMailBoxesRaw {
		if strings.Contains(MailBoxRaw, "LIST") {
			MailBoxRaw = strings.Replace(MailBoxRaw, "\"", "", -1)
			MailBoxPathSliceFirst := strings.Split(MailBoxRaw, ")")
			MailBoxPathSlice := strings.Split(MailBoxPathSliceFirst[1], " ")
			MailBoxPath := strings.Join(MailBoxPathSlice, "")
			AllMailBoxes = append(AllMailBoxes, MailBoxPath)
		}
	}
	if AllMailBoxesString == "" {
	}
	return AllMailBoxes
}

func ImapGetMailBoxInformation(conn *tls.Conn, Path string) *EVMailImapMailBoxInformation {
	if Path == "/" {
		log.Fatal("There is no Information available for the Server Root Directory")
	}
	TrimedPath := strings.TrimLeft(Path, "/")
	TrimedPath += "\r\n"
	log.Print(TrimedPath)
	RequestServerMessage := strings.Join([]string{".", "SELECT", TrimedPath}, " ")
	log.Print(RequestServerMessage)
	ImapSendServerMessage(conn, RequestServerMessage)
	ResponseMessage, _ := ImapReadServerMessage(conn)

	ImapMailBoxInfo := NewEVMailImapMailBoxInformation()

	if strings.Contains(ResponseMessage, "[NONEXISTENT]") {
		ImapMailBoxInfo.Flags = make([]string, 0)
		ImapMailBoxInfo.Mails = 0
		ImapMailBoxInfo.Recent = 0
		ImapMailBoxInfo.Path = Path
		ImapMailBoxInfo.NonExistent = true
		return ImapMailBoxInfo
	}

	ResponseMessageRaw := strings.Split(ResponseMessage, "\n")

	FlagsRaw := strings.Replace(ResponseMessageRaw[0], "* FLAGS (", "", -1)
	FlagsRaw = strings.Replace(FlagsRaw, ")", "", -1)
	ImapMailBoxInfo.Flags = strings.Split(FlagsRaw, " ")

	MailsRaw := strings.Split(ResponseMessageRaw[3], " ")
	ImapMailBoxInfo.Mails, _ = strconv.Atoi(MailsRaw[1])

	RecentRaw := strings.Split(ResponseMessageRaw[4], " ")
	ImapMailBoxInfo.Recent, _ = strconv.Atoi(RecentRaw[1])
	ImapMailBoxInfo.Path = Path
	ImapMailBoxInfo.NonExistent = false

	return ImapMailBoxInfo
}

func ImapGetMailBoxesInformation(conn *tls.Conn) *EVMailImapMailBoxes {
	MailBoxesInfo := NewEVMailImapMailBoxes()
	MailBoxes := ImapGetAllMailboxes(conn, "/")
	for _, MailBoxPath := range MailBoxes {
		MailBoxInfo := ImapGetMailBoxInformation(conn, MailBoxPath)
		MailBoxesInfo.MailBoxes = append(MailBoxesInfo.MailBoxes, *MailBoxInfo)
	}
	return MailBoxesInfo
}

func CheckRegExpString(Expression string, String string) string {
	RegExp, _ := regexp.Compile(Expression)
	match := RegExp.FindString(String)
	log.Printf("matched::::%v", match)
	return match
}

func ImapEmailSplitBoundary(EmailString string, conn *tls.Conn, MailId int) *EVMailEmail {
	log.Print(EmailString)
	MessageHeaderRaw := ImapGetRawMailBoxMessageHeader(conn, MailId)
	// todo: check if this 2 information are important
	//ImapParseMailHeader(MessageHeaderRaw, "Date: ")
	//ImapParseMailHeader(MessageHeaderRaw, "Content-Type: ")
	Mail := NewEVMailEmail()
	Mail.Server = "smtp.google.com"
	Mail.Port = 587
	Mail.User = "projects.notification@evalgo.com"
	Mail.Password = "IbeT2012"
	Mail.From = ImapParseMailHeader(MessageHeaderRaw, "From: ")
	Mail.To = []string{ImapParseMailHeader(MessageHeaderRaw, "To: ")}
	Mail.Subject = ImapParseMailHeader(MessageHeaderRaw, "Subject: ")
	Mail.Raw = EmailString
	MailBodyArray := strings.Split(EmailString, "\n")
	log.Printf("%v", MailBodyArray)
	Boundary := MailBodyArray[1]
	EmailsArray := strings.Split(EmailString, Boundary)
	log.Printf("BOUNDARY----------------%v", MailBodyArray[1])
	log.Printf("-----------------EMAILS %d :: %v", len(EmailsArray), EmailsArray)
	Mail.Body = EmailsArray[1]
	if len(EmailsArray) >= 3 {
		Mail.Html = EmailsArray[2]
	} else {
		Mail.Html = ""
	}
	//Mail.Attachments = make(map[string][]byte,0)
	return Mail
}

func ImapReadEmailMessage(conn *tls.Conn, MailId int) (*EVMailEmail, error) {
	replyHeader := make([]byte, 4096)
	replyString := ""
	for {
		sn, _ := conn.Read(replyHeader)
		replyMessage := string(replyHeader[:sn])
		replyString = strings.Join([]string{replyString, replyMessage}, "")
		if true == strings.Contains(replyString, strings.Join([]string{".", "OK", "Success"}, " ")) {
			// check for fetch imap response
			FetchCommand := CheckRegExpString(".+ FETCH \\(BODY\\[TEXT\\] {[0-9].+}", replyString)
			if FetchCommand != "" {
				replyString = strings.Replace(replyString, FetchCommand, "", -1)
			}
			// check for ending imap response success message
			SuccessCaseOne := CheckRegExpString("\015\012\\)\015\012. OK Success", replyString)
			if SuccessCaseOne != "" {
				replyString = strings.Replace(replyString, SuccessCaseOne, "", -1)
			}
			// check for ending imap response success message
			SuccessCaseTwo := CheckRegExpString("\\)\015\012. OK Success", replyString)
			if SuccessCaseTwo != "" {
				replyString = strings.Replace(replyString, SuccessCaseTwo, "", -1)
			}
			return ImapEmailSplitBoundary(replyString, conn, MailId), nil
		}
	}
	return ImapEmailSplitBoundary(replyString, conn, MailId), nil
}

func ImapConnect(Server string, Port int, Pem string, Key string) (*tls.Conn, string) {
	cert, err := tls.LoadX509KeyPair(Pem, Key)
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	ServerImapConnectionString := Server
	ServerImapConnectionString += ":"
	ServerImapConnectionString += strconv.Itoa(Port)
	conn, err := tls.Dial("tcp", ServerImapConnectionString, &config)
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}
	//log.Println("client: connected to: ", conn.RemoteAddr())
	state := conn.ConnectionState()
	//log.Printf("%v",state)
	for _, v := range state.PeerCertificates {
		fmt.Println("Client: Server public key is:")
		fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
	}
	//log.Println("client: handshake: ", state.HandshakeComplete)
	//log.Println("client: mutual: ", state.NegotiatedProtocolIsMutual)

	serverResponse, _ := ImapReadServerMessage(conn)
	log.Print(serverResponse)
	return conn, serverResponse
}

func ImapLogin(conn *tls.Conn, UserName string, Password string) string {
	// send login message
	LoginServerMessage := strings.Join([]string{".", " ", "LOGIN", " ", UserName, " ", Password, "\r\n"}, "")
	ImapSendServerMessage(conn, LoginServerMessage)
	loginResponse, _ := ImapReadServerMessage(conn)
	log.Print(loginResponse)
	return loginResponse
}

func ImapParseMailHeader(MessageHeaderRaw string, Header string) string {
	//!!!!!DO NOT DELETE THIS COMMENT BELOW!!!!!!!
	//RegularRule := "[0-9.+A-Z.+a-z.+].+\n" !!!!!
	//!!!!!----DO NOT DELETE THIS COMMENT ---!!!!!
	RegularRule := "(.*).+\n"

	Expression := strings.Join([]string{"\n", Header, RegularRule}, "")
	RegExp, _ := regexp.Compile(Expression)
	match := RegExp.FindString(MessageHeaderRaw)

	if Header == "Deliverd-To: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}

	if Header == "Received: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}

	if Header == "Return-Path: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}

	if Header == "Received-SPF: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}

	if Header == "Authentication-Results: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}

	if Header == "MIMVE-Version: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}

	if Header == "Date: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}

	if Header == "Message-ID: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}

	if Header == "Subject: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}

	if Header == "From: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}

	if Header == "To: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}

	if Header == "Content-Type: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}

	if Header == "X-Gm-Message-State: " {
		return strings.TrimLeft(strings.Replace(match, Header, "", -1), "\n")
	}
	return ""
}

func ImapGetRawMailBoxMessageHeader(conn *tls.Conn, MailId int) string {
	FetchServerMessage := strings.Join([]string{".", "FETCH", strconv.Itoa(MailId), "BODY[HEADER]\r\n"}, " ")
	log.Print(FetchServerMessage)
	ImapSendServerMessage(conn, FetchServerMessage)
	MessageHeaderRaw, _ := ImapReadMailBoxMessagesHeader(conn)
	return MessageHeaderRaw
}

func ImapGetMailBoxMessageHeader(conn *tls.Conn, MailId int) *EVMailImapEmailHeader {
	MessageHeaderRaw := ImapGetRawMailBoxMessageHeader(conn, MailId)
	HeaderInformation := NewEVMailImapEmailHeader()
	HeaderInformation.Id = MailId
	/* Avaiable Options *
	++++++++++++++++++++
	- HeaderInformations
	- Deliverd-To
	- Received
	- Return-Path
	- Received-SPF
	- Authentication-Results
	- MIME-Version
	- Date
	- Message-ID
	- Subject
	- From
	- To
	- Content-Type
	- X-Gm-Message-State
	*+++++++++++++++++++*/
	HeaderInformation.From = ImapParseMailHeader(MessageHeaderRaw, "From: ")
	HeaderInformation.To = ImapParseMailHeader(MessageHeaderRaw, "To: ")
	HeaderInformation.Subject = ImapParseMailHeader(MessageHeaderRaw, "Subject: ")
	HeaderInformation.Date = ImapParseMailHeader(MessageHeaderRaw, "Date: ")
	HeaderInformation.ContentType = ImapParseMailHeader(MessageHeaderRaw, "Content-Type: ")
	return HeaderInformation
}

func ImapGetAllMailBoxMessageHeader(conn *tls.Conn, MailBox *EVMailImapMailBoxInformation) (*EVMailImapMailBoxAllMessages, error) {
	MailBoxAllMessages := NewEVMailImapMailBoxAllMessages()
	for i := 1; i < (MailBox.Mails + 1); i++ {
		HeaderInformation := ImapGetMailBoxMessageHeader(conn, i)
		MailBoxAllMessages.Headers = append(MailBoxAllMessages.Headers, *HeaderInformation)
	}
	return MailBoxAllMessages, nil
}

func ImapGetMessageById(conn *tls.Conn, MailBoxPath string, MailId int) *EVMailImapEmail {
	MailBoxPath += "\r\n"
	SelectMailBoxServerMessage := strings.Join([]string{".", "SELECT", MailBoxPath}, " ")
	ImapSendServerMessage(conn, SelectMailBoxServerMessage)
	SelectResponse, _ := ImapReadServerMessage(conn)
	log.Print(SelectResponse)
	if SelectResponse == "" {
	}
	FetchServerMessage := strings.Join([]string{".", "FETCH", strconv.Itoa(MailId), "BODY[TEXT]\r\n"}, " ")
	log.Print(FetchServerMessage)
	ImapSendServerMessage(conn, FetchServerMessage)
	MessageObj, _ := ImapReadEmailMessage(conn, MailId)
	var ImapEmailMessage *EVMailImapEmail
	ImapEmailMessage = NewEVMailImapEmail()
	ImapEmailMessage.Id = MailId
	ImapEmailMessage.Body = MessageObj
	return ImapEmailMessage
}

func ImapLogout(conn *tls.Conn) string {
	ImapSendServerMessage(conn, ". LOGOUT\r\n")
	LogoutResponse, _ := ImapReadServerMessage(conn)
	conn.Close()
	return LogoutResponse
}
