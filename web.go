package main

import (
  "bytes"
  "crypto/hmac"
  "crypto/sha1"
	"fmt"
	"net/http"
	"os"
  "os/signal"
  "github.com/hoisie/web"
  "log"
  "time"
  "encoding/base64"
  "strconv"
  "strings"
)

type templateData struct {
  User User
  Info string
}

type homeData struct {
  templateData
}

type questionData struct {
  templateData
  Question Question
}

type loginData struct {
  templateData
}

type addQuestionData struct {
  templateData 
}

var (
  logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
)

func main() {

  // Configs
  web.Config.CookieSecret = os.Getenv("COOKIESECRET")

  // Handlers
  web.Get("/", homeHandler)
	web.Get("/quiz/", quizHandler)
  web.Get("/login/", loginHandler)
  web.Post("/login/", loginHandler)
  web.Get("/addquestion/", addQuestionHandler)
  web.Post("/addquestion/", addQuestionHandler)

  // Hooks to close db connections
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)
  go func() {
    for sig := range c {
      fmt.Printf("%+v\n", sig)
      CloseOpenSessions()
      os.Exit(1)
    }
  }()

  // Starting webserver
	location := fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
  web.Run(location)
}

func homeHandler(ctx *web.Context) {
  userid, _ := ctx.GetSecureCookie("userid")
  user, info := GetUserFromId(userid)
  logger.Println("userid: " + userid + ", " + fmt.Sprintf("%+v",user))
  Render("home", homeData{templateData{User: user, Info: info}}, ctx, ctx.Params["refresh"] != "")
}

func loginHandler(ctx *web.Context) {
  loginInfo := ""
  userid := ""
  if ctx.Params["username"] != "" {
    user := User{}
    user, loginInfo = LogIn(ctx.Params["username"], ctx.Params["password"])
    if loginInfo == "" {
      setCookie(ctx, "userid", user.UserId, 0)
      userid = user.UserId
    }
  }
  if userid == "" {
    userid, _ = ctx.GetSecureCookie("userid")
  }
  logger.Println(fmt.Sprintf("userid got from cookie: %+v", userid))
  loggedInUser, info := GetUserFromId(userid)
  Render("login", loginData{templateData{User: loggedInUser, Info: info}}, ctx, ctx.Params["refresh"] != "")
}

func addQuestionHandler(ctx *web.Context) {
  userid, _ := ctx.GetSecureCookie("userid")
  user, _ := GetUserFromId(userid)
  var info string

  logger.Println("quesiton: " + ctx.Params["question"] + info)
  if ctx.Params["question"] != "" {
    logger.Println("inside condition")
    var err error
    err, info = AddQuestion(Question{Question: ctx.Params["question"], Options: strings.Split(ctx.Params["options"], ";")})

    logger.Println(info)
    if err != nil {
      panic(err)
    }
  }

  Render("addquestion", addQuestionData{templateData{User: user, Info: info}}, ctx, ctx.Params["refresh"] != "")
}

func quizHandler(ctx *web.Context) {
  userid, _ := ctx.GetSecureCookie("userid")
  user, info := GetUserFromId(userid)
  Render("question", questionData{templateData{User: user, Info: info}, GetRandomQuestion()}, ctx, ctx.Params["refresh"] != "")
}

func setCookie(ctx *web.Context, name string, value string, age int64) {
  if len(ctx.Server.Config.CookieSecret) == 0 {
      ctx.Server.Logger.Println("Secret Key for secure cookies has not been set. Please assign a cookie secret to web.Config.CookieSecret.")
      return
  }
  var buf bytes.Buffer
  encoder := base64.NewEncoder(base64.StdEncoding, &buf)
  encoder.Write([]byte(value))
  encoder.Close()
  vs := buf.String()
  vb := buf.Bytes()
  timestamp := strconv.FormatInt(time.Now().Unix(), 10)
  sig := getCookieSig(ctx.Server.Config.CookieSecret, vb, timestamp)
  cookie := strings.Join([]string{vs, timestamp, sig}, "|")
  var expiry time.Time
  if age == 0 {
    expiry = time.Unix(2147483647, 0)
  } else {
    expiry = time.Unix(time.Now().Unix()+age, 0)
  }
  ctx.SetCookie(&http.Cookie{Name: name, Value: cookie, Expires: expiry, Path: "/"})
}

func getCookieSig(key string, val []byte, timestamp string) string {
  hm := hmac.New(sha1.New, []byte(key))

  hm.Write(val)
  hm.Write([]byte(timestamp))

  hex := fmt.Sprintf("%02x", hm.Sum(nil))
  return hex
}
