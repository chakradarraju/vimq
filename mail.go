package main

import (
  "net/smtp"
  "fmt"
  "text/template"
  "bytes"
)

var (
  server string = getenv("MAILSERVER")
  port string = getenv("MAILSERVERPORT")
  username string = getenv("MAILID")
  password string = getenv("MAILPASSWORD")
  auth smtp.Auth = smtp.PlainAuth("", username, password, server)
)

func renderMail(tmpl string, data interface{}) []byte {
  t, _ := template.ParseFiles("templates/" + tmpl + ".mail")
  var ret bytes.Buffer
  t.Execute(&ret, data)
  return ret.Bytes()
}

func mail(to []string, body []byte) {
  if err := smtp.SendMail(server + ":" + port, auth, username, to, body); err != nil {
    logger.Println("Error in sending mail: " + string(body) + "\n" + fmt.Sprintf("%+v", err))
  }
}
