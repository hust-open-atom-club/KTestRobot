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

var PATCH_DIR string
var BOOT_DIR string
var MAINLINE_DIR string
var LINUX_NEXT_DIR string
var SMATCH_DIR string
var KTBot_DIR string

func main() {
	BotInit()
	for {
		ReceiveEmail()
		time.Sleep(time.Minute * 20)
	}
}