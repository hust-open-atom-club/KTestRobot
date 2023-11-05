package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"
	"time"
)

type EmailHeader struct {
	FromName  string
	FromAddr  string
	Cc        []string
	MessageID string
	Subject   string
}

type Config struct {
	// Email account used to receive kernel patches, e.g., "ktestrobot@126.com"
	Username string `json:"username"`
	// Password used to log in
	Password string `json:"password"`
	// whitelist only to be processed
	WhiteLists []string `json:"whiteLists"`
}

type MailInfo struct {
	SMTPServer   string   `json:"smtpServer"`
	SMTPPort     int      `json:"smtpPort"`
	SMTPUsername string   `json:"smtpUsername"`
	SMTPPassword string   `json:"smtpPassword"`
	IMAPServer   string   `json:"imapServer"`
	IMAPPort     int      `json:"imapPort"`
	IMAPUsername string   `json:"imapUsername"`
	IMAPPassword string   `json:"imapPassword"`
	WhiteLists   []string `json:"whiteLists"`
}

type EmailConfig struct {
	SMTPServer string `json:"smtpServer"`
	SMTPPort   int    `json:"smtpPort"`
	IMAPServer string `json:"imapServer"`
	IMAPPort   int    `json:"imapPort"`
}

var (
	flagConfig = flag.String("config", "", "configuration file")
	//flagDebug  = flag.Bool("debug", false, "dump all the logs")
)

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

func parseConfig(configFile string) MailInfo {
	var config Config
	// open config file
	configFd, err := os.Open(configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer configFd.Close()
	// parse json file
	dec := json.NewDecoder(configFd)
	// disallow any unknown fields
	dec.DisallowUnknownFields()
	if err = dec.Decode(&config); err != nil {
		log.Fatal(err)
	}
	// retrieve email account configuration based on domain
	// TODO: add more configuration of email server
	mapDomaintoEmailAccount := map[string]EmailConfig{
		"126.com":     {"smtp.126.com", 25, "imap.126.com", 993},
		"hust.edu.cn": {"mail.hust.edu.cn", 465, "mail.hust.edu.cn", 993},
	}
	data := strings.Split(config.Username, "@")
	if len(data) != 2 {
		log.Fatalf("Please provide the correct email account, e.g., ktestrobot@126.com")
	}
	emailAccount, ok := mapDomaintoEmailAccount[data[1]]
	if !ok {
		log.Fatalf("We don't support domain %s", data[1])
	}
	mailinfo := MailInfo{
		emailAccount.SMTPServer,
		emailAccount.SMTPPort,
		config.Username,
		config.Password,
		emailAccount.IMAPServer,
		emailAccount.IMAPPort,
		config.Username,
		config.Password,
		config.WhiteLists,
	}
	return mailinfo
}

func main() {
	flag.Parse()
	if *flagConfig == "" {
		log.Fatalf("No config file specified")
	}
	mailinfo := parseConfig(*flagConfig)
	// init the environment of KTestRobot
	BotInit()
	for {
		mailinfo.ReceiveEmail()
		for _, patchname := range patchlist {
			checkresult := "--- Test Result ---\n"
			checkres := CheckPatchAll(patchname, ChangedPath)
			logname := patchname[:len(patchname)-6]
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
			mailinfo.SendEmail(toSend, emailheader)
		}
		patchlist = patchlist[:0]
		time.Sleep(time.Minute * 20)
	}
}
