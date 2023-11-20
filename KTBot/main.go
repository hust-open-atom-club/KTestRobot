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
	// Number of processes required to compile the kernel
	Procs 	 int `json:"procs"`
	//Time interval between two checks of the mail
	Interval int `json:"interval"`
	// whitelist only to be processed
	WhiteLists []string `json:"whiteLists"`
	//mail list for robot testing 
	MailingList string `json:"mailingList"`
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
	Procs 		 int      `json:"procs"`
	Interval     int      `json:"interval"`
	MailingList  string   `json:"mailingList"`
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
	if config.Procs == 0 {
		config.Procs = 20
	}
	if config.Interval == 0 {
		config.Interval = 20
	}
	if config.MailingList == "" {
		config.MailingList = "kernel_testing_robot@googlegroups.com"
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
		config.Procs,
		config.Interval,
		config.MailingList,
	}
	return mailinfo
}

func main() {
	flag.Parse()
	if *flagConfig == "" {
		log.Fatalf("No config file specified")
	}
	mailinfo := parseConfig(*flagConfig)
	// get current directory
	KTBot_DIR, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get the current directory", err)
	}
	// init the running environment of KTestRobot
	// patch stores the patches received from mailing list
	// log stores the log file
	// mainline stores the mainline source code
	// linux-next stores the linux-next source code
	// smatch stores the smatch source code
	mailinfo.botInit(KTBot_DIR)
	for {
		// receive emails from mailing list
		reader_list := mailinfo.ReceiveEmail(KTBot_DIR)
		if reader_list != nil {
			mailinfo.update(KTBot_DIR)
			// check all the received new emails and patches
			for _, mail_reader := range reader_list {
				// process the email and extract the original email sender and header
				// in this process, call checkers on the patch to check many aspects
				toSend, emailheader := mailinfo.MailProcess(mail_reader, KTBot_DIR)
				if toSend != "" {
					// send feedback emails to the sender
					mailinfo.SendEmail(toSend, emailheader)
				} else {
					continue
				}
			}
		}
		time.Sleep(time.Minute * time.Duration(mailinfo.Interval))
	}
}
