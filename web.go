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
  web.Get("/", simplePageHandler("home"))
  web.Get("/login/", simplePageHandler("login"))
  web.Get("/signup/", simplePageHandler("signup"))
  web.Get("/addquestion/", simplePageHandler("addquestion"))

	web.Get("/quiz/", quizHandler)

  web.Get("/logout/", logoutHandler)

  web.Post("/login/", loginSubmitHandler)
  web.Post("/signup/", signupSubmitHandler)
  web.Post("/addquestion/", addQuestionSubmitHandler)

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

func simplePageHandler(page string) func(*web.Context, ...alert) {
  return func(ctx *web.Context, alerts ...alert) {
    userid, _ := ctx.GetSecureCookie("userid")
    user, userAlerts := GetUserFromId(userid)
    alerts = append(alerts, userAlerts...)
    Render(page, templateData{User: user, Alerts: alerts}, ctx, ctx.Params["refresh"] != "")
  }
}

func quizHandler(ctx *web.Context, alerts ...alert) {
  userid, _ := ctx.GetSecureCookie("userid")
  user, userAlerts := GetUserFromId(userid)
  alerts = append(alerts, userAlerts...)
  Render("quiz", questionData{templateData:templateData{User: user, Alerts: alerts}, Question:GetRandomQuestion()}, ctx, ctx.Params["refresh"] != "")
}
func loginSubmitHandler(ctx *web.Context) {
  user, alerts := LogIn(ctx.Params["username"], ctx.Params["password"])
  if len(alerts) > 0 {
    simplePageHandler("login")(ctx, alerts...)
    return
  }
  setCookie(ctx, "userid", user.UserId, 0)
  ctx.Redirect(301, "/")
}

func logoutHandler(ctx *web.Context) {
  setCookie(ctx, "userid", "", -1) // Deleting cookie
  ctx.Redirect(301, "/")
}

func signupSubmitHandler(ctx *web.Context) {
  user, alerts := SignUp(ctx.Params["username"], ctx.Params["password"], ctx.Params["email"])
  if len(alerts) > 0 {
    simplePageHandler("signup")(ctx, alerts...)
    return
  }
  setCookie(ctx, "userid", user.UserId, 0)
  ctx.Redirect(301, "/")
}

func addQuestionSubmitHandler(ctx *web.Context) {
  var err error
  options := strings.Split(ctx.Params["options"], ";")
  correctoptionindex, err := strconv.Atoi(ctx.Params["correctoptionindex"])
  alerts := []alert{}
  if err != nil {
    alerts = append(alerts, alert{Text:"Problem finding correct option", Type: "danger"})
  } else {
    correctoption := options[correctoptionindex]
    alerts = append(alerts, AddQuestion(Question{Question: ctx.Params["question"], Options: options, CorrectOption: correctoption})...)
  }
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