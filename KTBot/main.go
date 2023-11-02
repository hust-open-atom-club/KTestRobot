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
var ChangedPath string
var emailheader EmailHeader
var patchlist []string
var LogMessage string


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
		for _, patchname := range patchlist{
			checkresult := "--- Test Result ---\n"
			checkres:= CheckPatchAll(patchname, ChangedPath)
			logname := patchname[:len(patchname) - 6]
			log_file, err2 := os.Create("log/" + logname)
			if err2 != nil {
				log.Println("open log_file: ", err2)
				return
			}
			defer log_file.Close()
			_, err3 := log_file.WriteString(checkres)
			if err3 != nil {
				log.Println("write log_file: ", err3)
				return
			}

			checkresult += checkres
			toSend := ChangedPath + LogMessage + checkresult
			SendEmail(toSend, emailheader)
		}
		patchlist = patchlist[:0]
		time.Sleep(time.Minute * 20)
	}
}
