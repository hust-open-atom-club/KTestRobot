package main

import (
	"time"
	//"sync"
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
var KTBot_DIR = "/home/lsc20011130/KTestRobot/KTBot/"

var username = "ktestrobot@126.com"
var passwd = "APSSXSSPWXLFXVUJ" //for ktestrobot@126.com

var StartTime = time.Date(2023, 4, 28, 0, 0, 0, 0, time.Local)
var enableStartTime = 0


func main() {
	BotInit()
	//go web()
	for {
		ReceiveEmail()
		/*
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			ReceiveEmail()
			wg.Done()
		}()
		go func() {
			ReadFile()
			wg.Done()
		}()
		wg.Wait()
		*/
		time.Sleep(time.Minute * 20)
	}
}
