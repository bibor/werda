package main

import (
	"bufio"
	"fmt"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

// struct for parsing an SSHD message from syslog
type SSHMsg struct {
	SystemDCGroup string `json:"_SYSTEMD_CGROUP"`
	SystemDUnit   string `json:"_SYSTEMD_UNIT"`
	Hostname      string `json:"_HOSTNAME"`
	Message       string `json:"Message"`
	Comm          string `json:"_COMM"`
	Timestamp     string `json:"SYSLOG_TIMESTAMP"`
}

// sends messages
type Sender interface {
	init() error
	send(*string) error
}

//dummy sender for debugging
type printSender struct {
}

// init function of debug  sender
// always nil
func (pS *printSender) init() error {
	return nil
}

// send function of debug sender
// prints message to stdout
// always nil
func (pS *printSender) send(msg *string) error {
	fmt.Println(*msg)
	return nil
}

// Sender for gotify backends
type gotifySender struct {
	Url      string
	AppToken string
}

// init function of the gotifySender
// gets gotify server and token from the environment variables
func (gS *gotifySender) init() error {
	fmt.Println("init of gotify sender")
	gS.Url = os.Getenv("GOTIFYSERVER")
	if len(gS.Url) == 0 {
		err := errors.New("GOTIFYSERVER environment variable is not defined") // ERROR
		return err
	}
	gS.AppToken = os.Getenv("GOTIFYTOKEN")
	if len(gS.AppToken) == 0 {
		err := errors.New("GOTIFYSERVER environment variable is not defined") // ERROR
		return err
	}
	fmt.Printf("Gotifyserver: %s\n", gS.Url)
	return nil
}

// send function for gotify sender
// takes msg  and sends it to configured backend
func (gS *gotifySender) send(msg *string) error {
	fmt.Printf("sending message with gotifySender to %s\n", gS.Url) // INFO
	urlWithToken := gS.Url + "message?token=" + gS.AppToken
	_, err := http.PostForm(urlWithToken,
		url.Values{"message": {*msg},
			"title":    {"New SSH Login"},
			"priority": {"5"}})
	return err
}

func main() {
	rawSyslogMsgChan := make(chan string)

	sender := gotifySender{}
	err := sender.init()
	if err != nil {
		fmt.Println(err)
		return
	}

	go extractSSHMsg(rawSyslogMsgChan, &sender)
	readSyslog(rawSyslogMsgChan)
}

// syslog reading loop
// reads syslog and pushes and forwards them to rawSyslogMsgChan
func readSyslog(rawSyslogMsgChan chan<- string) {
	syslogCmd := exec.Command("journalctl", "-o", "json", "-f")
	syslogOut, _ := syslogCmd.StdoutPipe()
	syslogReader := bufio.NewReader(syslogOut)
	syslogCmd.Start()
	syslogMsg, err := syslogReader.ReadString('\n')
	for err == nil {
		syslogMsg, err = syslogReader.ReadString('\n')
		rawSyslogMsgChan <- syslogMsg
	}
}

// gets a json object from syslog, exttracts the ssh message and forwards it to
// a sender
func extractSSHMsg(rawSyslogMsgChan <-chan string, sender Sender) {
	fmt.Println("started extraction loop")
	var err error = nil
	for err == nil {
		rawSyslogMsg := <-rawSyslogMsgChan
		msgJson := SSHMsg{}
		json.Unmarshal([]byte(rawSyslogMsg), &msgJson)
		// for debian it's ssh.service, for nixos it's sshd.service
		if msgJson.SystemDUnit == "sshd.service" {
			if strings.Contains(msgJson.Message, "Accepted publickey for") {
				fmt.Println("Publickey acceptance message detected") //INFO
				fmt.Println("sending messsage...")                   //INFO
				sMsg := fmt.Sprintf("New Login on %s: %s",
                    msgJson.Hostname,
                    msgJson.Message)
				go sender.send(&sMsg)
				fmt.Println("Message dispatched") //INFO
			}
		}
	}
}
