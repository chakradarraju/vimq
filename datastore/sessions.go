package datastore

import (
  "github.com/gorilla/sessions"
  "net/http"
)

var (
  store = sessions.NewFilesystemStore("")
)

type Session struct {
  session *sessions.Session
}

func (s Session) GetLoggedInUserId() string {
  if s.session.Values["userid"] == nil {
    return ""
  }
  return s.session.Values["userid"].(string)
}

func (s Session) Save(w http.ResponseWriter, r *http.Request) {
  s.session.Save(r, w)
}

func (s Session) SetLoggedInUser(u User) {
  s.session.Values["userid"] = u.UserId
  s.session.Values["username"] = u.UserName
}

func GetSession(sessionName string, r *http.Request) Session {
  sess, err := store.Get(r, sessionName)
  if err != nil {
    panic(err)
  }
  return Session{session:sess}
}
