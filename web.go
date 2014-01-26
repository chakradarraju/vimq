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

type alert struct {
  Text string
  Type string
}

type templateData struct {
  User User
  Alerts []alert
  PageTitle string
}

type homeData struct {
  templateData
}

type questionData struct {
  Question Question
  templateData
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
  web.Config.CookieSecret = getenv("COOKIESECRET")

  // Handlers
  web.Get("/", homeHandler)
	web.Get("/quiz/", quizHandler)
  web.Get("/login/", loginHandler)
  web.Post("/login/", loginSubmitHandler)
  web.Get("/logout/", logoutHandler)
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
	location := fmt.Sprintf("%s:%s", getenv("HOST"), getenv("PORT"))
  web.Run(location)
}

func getenv(env string) string {
  return os.Getenv(env)
}

func homeHandler(ctx *web.Context) {
  userid, _ := ctx.GetSecureCookie("userid")
  user, alerts := GetUserFromId(userid)
  Render("home", homeData{templateData{PageTitle: "VQuiz - Home", User: user, Alerts: alerts}}, ctx, ctx.Params["refresh"] != "")
}

func loginHandler(ctx *web.Context) {
  userid, _ := ctx.GetSecureCookie("userid")
  loggedInUser, alerts := GetUserFromId(userid)
  Render("login", loginData{templateData{PageTitle: "VQuiz - Login", User: loggedInUser, Alerts: alerts}}, ctx, ctx.Params["refresh"] != "")
}

func loginSubmitHandler(ctx *web.Context) {
  user, loginInfo := LogIn(ctx.Params["username"], ctx.Params["password"])
  if loginInfo == "" {
    setCookie(ctx, "userid", user.UserId, 0)
  }
  ctx.Redirect(301, "/")
}

func logoutHandler(ctx *web.Context) {
  setCookie(ctx, "userid", "", -1) // Deleting cookie
  ctx.Redirect(301, "/")
}

func addQuestionHandler(ctx *web.Context) {
  userid, _ := ctx.GetSecureCookie("userid")
  user, alerts := GetUserFromId(userid)

  if ctx.Params["question"] != "" {
    var err error
    options := strings.Split(ctx.Params["options"], ";")
    correctoptionindex, err := strconv.Atoi(ctx.Params["correctoptionindex"])
    if err != nil {
      alerts = append(alerts, alert{Text:"Problem finding correct option", Type: "error"})
    } else {
      correctoption := options[correctoptionindex]
      alerts = append(alerts, AddQuestion(Question{Question: ctx.Params["question"], Options: options, CorrectOption: correctoption})...)
    }
  }

  Render("addquestion", addQuestionData{templateData{PageTitle: "VQuiz - Add Question", User: user, Alerts: alerts}}, ctx, ctx.Params["refresh"] != "")
}

func quizHandler(ctx *web.Context) {
  userid, _ := ctx.GetSecureCookie("userid")
  user, alerts := GetUserFromId(userid)
  qdata := questionData{templateData:templateData{PageTitle: "VQuiz", User: user, Alerts: alerts}, Question:GetRandomQuestion()}
  Render("question", qdata, ctx, ctx.Params["refresh"] != "")
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
