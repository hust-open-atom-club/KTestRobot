package main

import (
	"time"
)

type EmailHeader struct {
	FromName  string
	FromAddr  string
	Cc        []string
	MessageID string
	Subject   string
}

var PATCH_DIR = "/path/to/patchdir/" //can be set automatically in BotInit()
var BUILD_DIR = "/path/to/build/linux/"
var BOOT_DIR = "/path/to/bootdir/"
var MAINLINE_DIR = "/home/lsc20011130/linux/"
var LINUX_NEXT_DIR = "/home/lsc20011130/linux-next/linux-next-next-20230609/"
var SMATCH_DIR = "/home/lsc20011130/smatch/"
var KTBot_DIR = "/home/lsc20011130/robot/KTBot/"

var username = "ktestrobot@126.com"
var passwd = "APSSXSSPWXLFXVUJ" //for ktestrobot@126.com

func main() {
	BotInit()
	for {
		ReceiveEmail()
		time.Sleep(time.Minute * 20)
	}
}
