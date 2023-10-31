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
	// 读取配置文件
	configFile, err := os.Open("config.json")
	if err != nil {
	log.Fatal(err)
	}
	defer configFile.Close()
	
	// 解析配置文件
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
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
