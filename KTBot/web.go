package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func indexPage(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html", "template.html")
	var v TBug
	v.Title = "KTestRobot"

	v.PatchBugs, v.OtherBugs = getBugList()
	// log.Println(v)
	t.Execute(w, v)
}

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006/01/02 15:04")
}

func UploadPatch(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("upload.html")
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		file, handler, err := r.FormFile("file")
		if err != nil {
			log.Println("UploadPatch: Error Retrieving the File")
			log.Println(err)
			return
		}
		defer file.Close()
		if !strings.HasSuffix(handler.Filename, ".patch") {
			http.Error(w, ".patch file only!", http.StatusBadRequest)
			return
		}

		//create file
		filename := strings.ReplaceAll(handler.Filename, " ", "")
		patch := "upload" + time.Now().Format("20060102150405") + "_" + filename
		dst, err := os.Create("uploadpatch/" + patch)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		//write to the file
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

func web() {

	myMux := http.NewServeMux()
	myMux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	myMux.Handle("/patch/", http.StripPrefix("/patch/", http.FileServer(http.Dir("patch/"))))
	myMux.HandleFunc("/upload", UploadPatch)
	myMux.HandleFunc("/index", indexPage)

	log.Println("Listening at 9090...")
	http.ListenAndServe(":9090", myMux)

}
