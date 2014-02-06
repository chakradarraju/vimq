package main

import (
  "github.com/hoisie/web"
  "encoding/json"
  "os"
  "strings"
  "strconv"
  "log"
  "github.com/gorilla/sessions"
)

var (
  store *sessions.FilesystemStore = sessions.NewFilesystemStore("", []byte(getenv("COOKIESECRET")))
  logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
)

func encodeJson(data interface{}) ([]byte, error) {
  return json.Marshal(data)
}

func getenv(env string) string {
  return os.Getenv(env)
}

func getLoggedInUser(ctx *web.Context) User {
  userid, _ := ctx.GetSecureCookie("userid")
  if userid == "" {
    return User{}
  }
  return GetUserFromId(userid, getNotifier(ctx))
}

func getNotifier(ctx *web.Context) func(string,string) {
  session, _ := store.Get(ctx.Request, "session")
  return func(typ string, message string) {
    if _, ok := session.Values[typ]; !ok {
      session.Values[typ] = []string{}
    }
    session.Values[typ] = append(session.Values[typ].([]string), message)
    session.Save(ctx.Request, ctx)
  }
}

func getNotifications(ctx *web.Context) map[string][]string {
  session, _ := store.Get(ctx.Request, "session")
  ret := map[string][]string{}
  for k, v := range session.Values {
    ret[k.(string)] = v.([]string)
    delete(session.Values, k)
  }
  session.Save(ctx.Request, ctx)
  return ret
}

func getOptions(optionsstr string, correctoptionindexstr string, notifier func(string,string)) ([]string, string) {
  options := strings.Split(optionsstr, ";")
  correctoptionindex, err := strconv.Atoi(correctoptionindexstr)
  if err != nil || correctoptionindex >= len(options) {
    notifier("danger", "Problem finding correct option")
    return []string{}, ""
  }
  return options, options[correctoptionindex]
}
