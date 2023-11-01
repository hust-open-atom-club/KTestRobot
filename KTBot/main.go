package main

import (
	"time"
	"os"
	"log"
	"encoding/json"
)

type EmailHeader struct {
	FromName  string
	FromAddr  string
	Cc        []string
	MessageID string
	Subject   string
}

type Config struct {
	SMTPServer   string `json:"smtpServer"`
	SMTPPort     int    `json:"smtpPort"`
	SMTPUsername string `json:"smtpUsername"`
	SMTPPassword string `json:"smtpPassword"`
	IMAPServer   string `json:"imapServer"`
	IMAPPort     int    `json:"imapPort"`
	IMAPUsername string `json:"imapUsername"`
	IMAPPassword string `json:"imapPassword"`
	WhiteLists []string `json:"whiteLists"`
}

var config Config
var PATCH_DIR string
var BOOT_DIR string
var MAINLINE_DIR string
var LINUX_NEXT_DIR string
var SMATCH_DIR string
var KTBot_DIR string


func init() {
	// open config file
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()
	
	// parse json file
	dec := json.NewDecoder(configFile)
	// disallow any unknown fields
	dec.DisallowUnknownFields()

	if err = dec.Decode(&config); err != nil {
		log.Fatal(err)
	}
}

func main() {
	BotInit()
	for {
		ReceiveEmail()
		time.Sleep(time.Minute * 20)
	}
}
