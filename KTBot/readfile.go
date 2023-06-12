package main

import (
	"log"
	//"os/exec"
	"strings"
	"os"
	//"bufio"
)

/*
func ReadFile() {
	infos, err := os.ReadDir("uploadpatch/")
	if err != nil {
		log.Println("ReadDir err: ", err)
		return
	}
	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name()
	}
	
	for _, patch := range names {
		file, err := os.OpenFile("uploadpatch/"+patch, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Println("Readfile: ", err)
			continue
		}
		log.Println("Readfile: ", patch)
		defer file.Close()
		buf := bufio.NewScanner(file)
		changedpath := ""
		flag := 0
		for {
			if !buf.Scan() {
				break
			}
			line := buf.Text()
			if strings.HasPrefix(line, "diff --git a") {
				flag = 1
				subline := line[13:]
				tempindex := strings.Index(subline, " ")
				changedpath += subline[:tempindex] + "\n"
			}
		}
		if flag == 0 {
			rmfile := exec.Command("rm", "uploadpatch/"+patch)
			rmfile.Run()
			continue
		} else {
			mvfile := exec.Command("mv", "uploadpatch/"+patch, "./patch/")
			mvfile.Run()
		}
		res, csvres := CheckPatchAll(patch, changedpath)
		
		Changed := "--- Changed Paths ---\n" + changedpath
		checkres := "\n--- Test Result ---\n" + res
		SendEmail(Changed + checkres, *defmailrecv)
		ResultProcess(res, csvres, patch)
	}
}
*/

func ResultProcess(checkres string, csvres string, patchname string) {
	writecsv := patchname + csvres + "\n"
	csvfile, csverr := os.OpenFile("res.csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if csverr != nil {
		log.Println("openfile: ", csverr)
		return
	}
	defer csvfile.Close()
	_, cerr := csvfile.WriteString(writecsv)
	if cerr != nil {
		log.Println("write csvfile: ", cerr)
		return
	}
	patchlog := strings.ReplaceAll(patchname, ".patch", ".log")
	reslog, logerr := os.Create("log/" + patchlog)
	if logerr != nil {
		log.Println("openfile: ", logerr)
		return
	}
	defer reslog.Close()
	_, lerr := reslog.WriteString(checkres)
	if lerr != nil {
		log.Println("write csvfile: ", lerr)
		return
	}
}


func BotInit() bool {
	dir, err := os.Getwd()
	if err != nil {
		log.Println("Init: ", err)
		return false
	}
	PATCH_DIR = dir + "/patch/"
	os.MkdirAll("./patch", 0777)
	os.MkdirAll("./log", 0777)
	//os.MkdirAll("./uploadpatch", 0777)
	return true
}