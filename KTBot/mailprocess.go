package main

import (
	"io"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"path/filepath"

	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

func (mailinfo MailInfo) CheckWhiteLists(mailaddr string) bool {
	var flag bool // default is false
	for _, suffix := range mailinfo.WhiteLists {
		if strings.Contains(mailaddr, suffix) {
			flag = true
			break
		}
	}
	return flag
}

func (mailinfo MailInfo) SendEmail(toSend string, emailheader EmailHeader) {
	//if pass all check, just send to patch committer
	mailtext := "Hi, " + emailheader.FromName + "\n\n"
	mailtext += "This email is automatically replied by KTestRobot(version 1.0). "
	mailtext += "Please do not reply to this email.\n"
	mailtext += "If you have any questions or suggestions about KTestRobot, "
	mailtext += "please contact Lishuchang <U202011978@hust.edu.cn>\n\n"
	mailtext += toSend
	mailtext += "\n-- \nKTestRobot(version 1.0)"
	log.Println("Connecting to smtp server")
	to := []string{emailheader.FromAddr}
	to = append(to, mailinfo.MailingList)
	msg := []byte("To: " + emailheader.FromAddr + "\r\n" +
		"Subject: Re: " + emailheader.Subject + "\r\n" +
		"Cc: " + strings.Join(emailheader.Cc, ";") + "\r\n" +
		"In-Reply-To: " + "<" + emailheader.MessageID + ">" + "\r\n" +
		"\r\n" +
		mailtext + "\r\n")
	auth := smtp.PlainAuth("", mailinfo.SMTPUsername, mailinfo.SMTPPassword, mailinfo.SMTPServer)
	err := smtp.SendMail(mailinfo.SMTPServer+":"+strconv.Itoa(mailinfo.SMTPPort), auth, mailinfo.SMTPUsername, to, msg)
	if err != nil {
		log.Println("SendMail", err)
	}
	log.Println("Successfully Send to: ", to)
}

func (mailinfo MailInfo) ReceiveEmail(KTBot_DIR string) []*mail.Reader {
	log.Println("Connecting to server...")
	// Connect to server
	c, err := client.DialTLS(mailinfo.IMAPServer+":"+strconv.Itoa(mailinfo.IMAPPort), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")
	idClient := id.NewClient(c)
	idClient.ID(
		id.ID{
			id.FieldName:    "KTestRobot",
			id.FieldVersion: "1.0.0",
		},
	)

	defer c.Logout()

	if err := c.Login(mailinfo.IMAPUsername, mailinfo.IMAPPassword); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()
	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	from := uint32(1)
	to := mbox.Messages
	if mbox.Recent == 0 {
		log.Println("No New Email")
		return nil
	}
	from = mbox.Messages - mbox.Recent + 1

	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, section.FetchItem()}
	messages := make(chan *imap.Message, to-from+2)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	var reader_list []*mail.Reader
	for msg := range messages {
		// check subject
		if !strings.HasPrefix(msg.Envelope.Subject, "[PATCH") {
			continue
		}

		section, err := imap.ParseBodySectionName("BODY[]")
		if err != nil {
			log.Println("ParseBodySectionName err!")
			continue
		}
		r := msg.GetBody(section)
		if r == nil {
			continue
		}

		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Println("CreateReader fail: ", err)
			continue
		}
		reader_list = append(reader_list, mr)
	}
	return reader_list
}

func (mailinfo MailInfo) MailProcess(mr *mail.Reader, KTBot_DIR string) (toSend string, h EmailHeader) {
	header := mr.Header
	var emailheader EmailHeader
	var patchname string
	var mailtext string
	var changed_path string
	var ignore bool // default is false
	var flag bool   // default is false
	if from, err := header.AddressList("From"); err == nil {
		log.Println("From:", from)
		name := ""
		if from[0].Name == name {
			index := strings.Index(from[0].Address, "@")
			name = from[0].Address[:index]
		} else {
			name = from[0].Name
		}
		patchname = strings.ReplaceAll(name, " ", "") + "_"
		emailheader.FromName = name
		emailheader.FromAddr = from[0].Address
	}
	if date, err := header.Date(); err == nil {
		log.Println("Date: ", date)
		date = date.Local()
		patchname += date.Format("20060102150405") + ".patch"
	}
	if subject, err := header.Subject(); err == nil {
		log.Println("Subject:", subject)
		emailheader.Subject = subject
	}
	if cclist, err := header.AddressList("Cc"); err == nil {
		log.Println("Cc:", cclist)
		for _, cc := range cclist {
			if mailinfo.CheckWhiteLists(cc.Address) {
				emailheader.Cc = append(emailheader.Cc, cc.Address)
			} else {
				ignore = true
				break
			}
		}
	}
	if to, err := header.AddressList("To"); err == nil {
		log.Println("To: ", to)
		for _, cc := range to {
			if mailinfo.CheckWhiteLists(cc.Address) {
				emailheader.Cc = append(emailheader.Cc, cc.Address)
			} else {
				ignore = true
				break
			}
		}
	}
	if ignore {
		log.Println("The email was not sent to internal, ignore.")
		return
	}
	if msgid, err := header.MessageID(); err == nil {
		log.Println("MessageID: ", msgid)
		emailheader.MessageID = msgid
	}

	//process the txt of mail
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(err)
			continue
		}
		// This is the message's text (can be plain-text or HTML)
		b, _ := io.ReadAll(p.Body)
		log.Println("Got text: \n", string(b))
		mailtext = string(b)
	}

	mailsplit := strings.Split(mailtext, "\n")
	ChangedPath := "--- Changed Paths ---\n"
	LogMessage := ""
	for _, line := range mailsplit {
		if strings.HasPrefix(line, "diff --git a") {
			flag = true
			subline := line[13:]
			tempindex := strings.Index(subline, " ")
			changed_path = subline[:tempindex]
			ChangedPath += subline[:tempindex] + "\n"
		} else if strings.HasPrefix(line, "Reviewed-by:") {
			log.Println("The patch has Reviewed-by tag.")
			return
		}
	}
	if flag {
		MessageEnd := strings.Index(mailtext, "Fixes:")
		if MessageEnd == -1 {
			MessageEnd = strings.Index(mailtext, "Signed-off-by:")
		}
		if MessageEnd == -1 {
			return
		}
		if MessageEnd > 0 {
			LogMessage = "\n--- Log Message ---\n"
			LogMessage += mailtext[:MessageEnd-1] + "\n"
		}

		var patch string
		index := strings.Index(mailtext, "You received this message because")
		if index == -1 {
			patch = mailtext
		} else {
			patch = mailtext[:index-5]
		}
		file, err := os.Create(filepath.Join("patch", patchname))
		if err != nil {
			log.Println("openfile: ", err)
			return
		}
		defer file.Close()
		patchheader := "From: " + emailheader.FromName
		patchheader += "<" + emailheader.FromAddr + ">\n"
		patchheader += "Subject: " + emailheader.Subject + "\n\n"
		_, err1 := file.WriteString(strings.ReplaceAll(patchheader+patch, "\r\n", "\n"))
		if err1 != nil {
			log.Println("write file: ", err1)
			return
		}

		checkresult := "--- Test Result ---\n"
		checkres := mailinfo.CheckPatchAll(KTBot_DIR, patchname, changed_path)

		logname := patchname[:len(patchname)-6]
		log_file, err2 := os.Create(filepath.Join("log", logname))
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
		return toSend, emailheader
	} else {
		log.Println("No Patch in this mail!")
		var empty EmailHeader
		return "", empty
	}
}
