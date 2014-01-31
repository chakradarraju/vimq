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
  Context *web.Context
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

type profileData struct {
  templateData
  Profile User
  AddedQuestions []Question
}

type editQuestionData struct {
  templateData
  Question Question
}

var (
  logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
)

func main() {

  // Configs
  web.Config.CookieSecret = getenv("COOKIESECRET")

  // Handlers
  web.Get("/", func(ctx *web.Context) { ctx.Redirect(301, "/home/") })
  web.Get("/home/", simplePageHandler("home"))
  web.Get("/login/", simplePageHandler("login"))
  web.Get("/signup/", simplePageHandler("signup"))
  web.Get("/addquestion/()", editQuestionHandlerGen(false))
  web.Get("/myprofile/()", profileHandler)
  web.Get("/profile/(.*)", profileHandler)
  web.Get("/question/(.*)/edit/",editQuestionHandlerGen(false))
  web.Get("/question/(.*)/", questionHandler)

	web.Get("/quiz/()", questionHandler)

  web.Get("/logout/", logoutHandler)

  web.Post("/login/", loginSubmitHandler)
  web.Post("/signup/", signupSubmitHandler)
  web.Post("/addquestion/", addQuestionSubmitHandler)
  web.Post("/question/(.*)/edit/", editQuestionHandlerGen(true))

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

func isActiveTab(base string, ctx *web.Context) bool {
  return strings.Index(ctx.Request.URL.Path, base) == 0
}

func getenv(env string) string {
  return os.Getenv(env)
}

func getLoggedInUser(ctx *web.Context) (User, []alert) {
  userid, _ := ctx.GetSecureCookie("userid")
  if userid == "" {
    return User{}, []alert{}
  }
  return GetUserFromId(userid)
}

func simplePageHandler(page string) func(*web.Context, ...alert) {
  return func(ctx *web.Context, alerts ...alert) {
    user, userAlerts := getLoggedInUser(ctx)
    alerts = append(alerts, userAlerts...)
    Render(page, templateData{User: user, Alerts: alerts, Context: ctx}, ctx, ctx.Params["refresh"] != "")
  }
}

func profileHandler(ctx *web.Context, userId string, alerts ...alert) {
  loggedInUser, alerts := getLoggedInUser(ctx)
  var user User
  if len(userId) == 0 {
    user = loggedInUser
  } else {
    var userAlerts []alert
    user, userAlerts = GetUserFromUserName(userId)
    alerts = append(alerts, userAlerts...)
  }
  addedQuestions, questionAlerts := getQuestionsFromId(user.AddedQuestionIds)
  alerts = append(alerts, questionAlerts...)
  Render("profile", profileData{templateData:templateData{User: loggedInUser, Alerts:alerts, Context: ctx}, Profile: user, AddedQuestions: addedQuestions}, ctx, ctx.Params["refresh"] != "")
}

func editQuestionHandlerGen(save bool) func(*web.Context, string, ...alert) {
  return func(ctx *web.Context, questionId string, alerts ...alert) {
    user, userAlerts := getLoggedInUser(ctx)
    alerts = append(alerts, userAlerts...)
    question := Question{}
    if len(questionId) > 0 {
      questionAlerts := []alert{}
      question, questionAlerts = getQuestionFromId(questionId)
      alerts = append(alerts, questionAlerts...)
      if question.AddedUserId != user.UserId {
        alerts = append(alerts, alert{Text:"Question was added by different user, you can't edit it.", Type:"danger"})
        Render("empty", templateData{User: user, Alerts: alerts, Context: ctx}, ctx, ctx.Params["refresh"] != "")
        return
      }
      if save {
        options, correctoption, alerts := getOptions(ctx.Params["options"], ctx.Params["correctoptionindex"])
        question.Question = ctx.Params["question"]
        question.Options = options
        question.CorrectOption = correctoption
        question.Save()
        alerts = append(alerts, alert{Text:"Question saved successfully", Type: "success"})
      }
    }
    Render("editquestion", editQuestionData{templateData:templateData{User: user, Alerts: alerts, Context: ctx}, Question: question}, ctx, ctx.Params["refresh"] != "")
  }
}

func questionHandler(ctx *web.Context, questionId string, alerts ...alert) {
  user, userAlerts := getLoggedInUser(ctx)
  alerts = append(alerts, userAlerts...)
  question := Question{}
  if len(questionId) > 0 {
    questionAlerts := []alert{}
    question, questionAlerts = getQuestionFromId(questionId)
    alerts = append(alerts, questionAlerts...)
  } else {
    question = getRandomQuestion()
  }
  Render("question", questionData{templateData:templateData{User: user, Alerts: alerts, Context: ctx}, Question:question}, ctx, ctx.Params["refresh"] != "")
}

func loginSubmitHandler(ctx *web.Context) {
  user, alerts := LogIn(ctx.Params["username"], ctx.Params["password"])
  if len(alerts) > 0 {
    simplePageHandler("login")(ctx, alerts...)
    return
  }
  setCookie(ctx, "userid", user.UserId, 0)
  ctx.Redirect(301, "/home/")
}

func logoutHandler(ctx *web.Context) {
  setCookie(ctx, "userid", "", -1) // Deleting cookie
  ctx.Redirect(301, "/home/")
}

func signupSubmitHandler(ctx *web.Context) {
  user := User {
    UserId: GetNextId("user"),
    UserName: ctx.Params["username"],
    DisplayName: ctx.Params["displayname"],
    PassWord: ctx.Params["password"],
    EmailId: ctx.Params["email"],
    UserLevel: "Rookie",
  }
  var alerts []alert
  user, alerts = SignUp(user)
  if len(alerts) > 0 {
    simplePageHandler("signup")(ctx, alerts...)
    return
  }
  setCookie(ctx, "userid", user.UserId, 0)
  ctx.Redirect(301, "/home/")
}

func getOptions(optionsstr string, correctoptionindexstr string) ([]string, string, []alert) {
  options := strings.Split(optionsstr, ";")
  correctoptionindex, err := strconv.Atoi(correctoptionindexstr)
  alerts := []alert{}
  if err != nil || correctoptionindex >= len(options) {
    return []string{}, "", []alert{alert{Text:"Problem finding correct option", Type: "danger"}}
  }
  return options, options[correctoptionindex], alerts
}

func addQuestionSubmitHandler(ctx *web.Context) {
  loggedInUser, _ := getLoggedInUser(ctx)
  options, correctoption, alerts := getOptions(ctx.Params["options"], ctx.Params["correctoptionindex"])
  if len(options) > 0 {
    question := Question {
      QuestionId: GetNextId("question"),
      Question: ctx.Params["question"],
      Options: options,
      CorrectOption: correctoption,
      AddedUserId:loggedInUser.UserId,
    }
    alerts = append(alerts, AddQuestion(question)...)
    ctx.Redirect(301, "/question/" + question.QuestionId + "/edit/")
  }
  simplePageHandler("editquestion")(ctx, alerts...)
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
