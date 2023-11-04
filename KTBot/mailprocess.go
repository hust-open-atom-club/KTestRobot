package main

import (
	"log"
	"strings"
	"io"
	"net/smtp"
	"strconv"
	"os"
	"github.com/emersion/go-imap"
	id "github.com/emersion/go-imap-id"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

func (mailinfo MailInfo) SendEmail(result string, h EmailHeader) {
	//if pass all check, just send to patch committer
	mailtext := "Hi, " + h.FromName + "\n"
	mailtext += "This email is automatically replied by KTestRobot(Beta). "
	mailtext += "Please do not reply to this email.\n"
	mailtext += "If you have any questions or suggestions about KTestRobot, "
	mailtext += "please contact Lishuchang <U202011978@hust.edu.cn>\n\n"
	mailtext += result
	mailtext += "\n-- \nKTestRobot(Beta)"
	log.Println("Connecting to smtp server")
	to := []string{h.FromAddr}
	msg := []byte("To: " + h.FromAddr + "\r\n" +
		"Subject: Re: " + h.Subject + "\r\n" +
		"Cc: " + strings.Join(h.Cc, ";") + "\r\n" +
		"In-Reply-To: " + "<" + h.MessageID + ">" + "\r\n" +
		"\r\n" +
		mailtext + "\r\n")
	auth := smtp.PlainAuth("", mailinfo.SMTPUsername, mailinfo.SMTPPassword, mailinfo.SMTPServer)
	err := smtp.SendMail(mailinfo.SMTPServer + ":" + strconv.Itoa(mailinfo.SMTPPort), auth, mailinfo.SMTPUsername, to, msg)
	if err != nil {
		log.Println("SendMail", err)
	}
	log.Println("Successfully Send to: ", to)
}

func (mailinfo MailInfo) ReceiveEmail() {
	log.Println("Connecting to server...")
	// Connect to server
	c, err := client.DialTLS(mailinfo.IMAPServer + ":" + strconv.Itoa(mailinfo.IMAPPort), nil)
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
		return
	}
	from = mbox.Messages - mbox.Recent + 1

	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, section.FetchItem()}
	messages := make(chan *imap.Message, to - from + 2)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

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
		header := mr.Header
		var patchname string
		var ignore = 0
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
				if WhiteLists(cc.Address, mailinfo) == 1 {
					emailheader.Cc = append(emailheader.Cc, cc.Address)
				} else {
					ignore = 1
					break
				}
			}
		}
		if to, err := header.AddressList("To"); err == nil {
			log.Println("To: ", to)
			for _, cc := range to {
				if WhiteLists(cc.Address, mailinfo) == 1 {
					emailheader.Cc = append(emailheader.Cc, cc.Address)
				} else {
					ignore = 1
					break
				}
			}
		}
		if ignore == 1 {
			log.Println("The email was not sent to internal, ignore.")
			continue
		}
		if msgid, err := header.MessageID(); err == nil {
			log.Println("MessageID: ", msgid)
			emailheader.MessageID = msgid
		}
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				// log.Fatal(err)
				log.Println(err)
				continue
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				// This is the message's text (can be plain-text or HTML)
				b, _ := io.ReadAll(p.Body)
				log.Println("Got text: \n", string(b))
				MailProcess(string(b), patchname, emailheader)
			case *mail.AttachmentHeader:
				// This is an attachment
				filename, _ := h.Filename()
				log.Println("Got attachment: \n", filename)
			}
		}
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}

func WhiteLists(mailaddr string, mailinfo MailInfo) int {
	var flag = 0
	for _, suffix := range mailinfo.WhiteLists {
		if strings.Contains(mailaddr, suffix) {
			flag = 1
			break
		}
	}
	return flag
}

func MailProcess(mailtext string, patchname string, h EmailHeader) {
	mailsplit := strings.Split(mailtext, "\n")
	var flag int = 0
	ChangedPath = "--- Changed Paths ---\n"
	LogMessage = ""
	for _, line := range mailsplit {
		if strings.HasPrefix(line, "diff --git a") {
			flag = 1
			subline := line[13:]
			tempindex := strings.Index(subline, " ")
			ChangedPath += subline[:tempindex] + "\n"
		} else if strings.HasPrefix(line, "Reviewed-by:") {
			log.Println("The patch has Reviewed-by tag.")
			return
		}
	}
	if flag == 1 {
		MessageEnd := strings.Index(mailtext, "Fixes:")
		if MessageEnd == -1 {
			MessageEnd = strings.Index(mailtext, "Signed-off-by:")
		}
		if MessageEnd == -1 {
			return
		}
		if MessageEnd > 0 {
			LogMessage = "\n--- Log Message ---\n"
			LogMessage += mailtext[:MessageEnd - 1] + "\n"
		}
		
		var patch string
		index := strings.Index(mailtext, "You received this message because")
		if index == -1 {
			patch = mailtext
		} else {
			patch = mailtext[:index-5]
		}
		file, err := os.Create("patch/" + patchname)
		if err != nil {
			log.Println("openfile: ", err)
			return
		}
		defer file.Close()
		patchheader := "From: " + h.FromName
		patchheader += "<" + h.FromAddr + ">\n"
		patchheader += "Subject: " + h.Subject + "\n\n"
		_, err1 := file.WriteString(strings.ReplaceAll(patchheader + patch, "\r\n", "\n"))
		if err1 != nil {
			log.Println("write file: ", err1)
			return
		}

		//cmd := exec.Command("fromdos", PATCH_DIR + patchname)
		//cmderr := cmd.Run()
		//if cmderr != nil {
		//	log.Println("fromdos: ", cmderr)
		//}
		patchlist = append(patchlist, patchname)
		// checkresult := "--- Test Result ---\n"
		// checkres:= CheckPatchAll(patchname, ChangedPath)

		// logname := patchname[:len(patchname) - 6]
		// log_file, err2 := os.Create("log/" + logname)
		// if err2 != nil {
		// 	log.Println("open log_file: ", err2)
		// 	return
		// }
		// defer log_file.Close()
		// _, err3 := log_file.WriteString(checkres)
		// if err3 != nil {
		// 	log.Println("write log_file: ", err3)
		// 	return
		// }

		// checkresult += checkres
		// toSend := ChangedPath + LogMessage + checkresult
		// SendEmail(toSend, h)
	} else {
		log.Println("No Patch in this mail!")
	}
}
