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
	KTBot_DIR, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get the current directory", err)
	}
	// init the environment of KTestRobot
	// patch stores the patches received from mailing list
	// log stores the log file
	// mainline stores the mainline source code
	// linux-next stores the linux-next source code
	// smatch stores the smatch source code
	botInit(KTBot_DIR)
	for {
		mailinfo.ReceiveEmail(KTBot_DIR)
		time.Sleep(time.Minute * 20)
	}
}
