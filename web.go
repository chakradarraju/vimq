package main

import (
	"fmt"
	"os"
  "os/signal"
  "github.com/hoisie/web"
  "log"
  "strconv"
  "strings"
)

type templateData struct {
  User User
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

/*
   TODOs:
   * ajax posts
   * server validation
   * replace cookie based alerts
   * implement returnto for request that will redirect
*/

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
  web.Get("/profile/(.*)/", profileHandler)
  web.Get("/question/(.*)/edit/",editQuestionHandlerGen(false))
  web.Get("/question/(.*)/", questionHandler)

	web.Get("/quiz/()", questionHandler)

  web.Get("/logout/", logoutHandler)

  web.Post("/login/", loginSubmitHandler)
  web.Post("/signup/", signupSubmitHandler)
  web.Post("/addquestion/", addQuestionSubmitHandler)
  web.Post("/question/(.*)/edit/", editQuestionHandlerGen(true))
  web.Post("/question/(.*)/delete/", deleteQuestionHandler)

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

func getLoggedInUser(ctx *web.Context) User {
  userid, _ := ctx.GetSecureCookie("userid")
  if userid == "" {
    return User{}
  }
  return GetUserFromId(userid, getNotifier(ctx))
}

//var notifications map[*web.Context]map[string][]string

func getNotifier(ctx *web.Context) func(string,string) {
  return func(typ string, message string) {
    appendToCookie(ctx, typ, message, 30)
  }
}

func simplePageHandler(page string) func(*web.Context) {
  return func(ctx *web.Context) {
    user := getLoggedInUser(ctx)
    Render(page, templateData{User: user, Context: ctx}, ctx, ctx.Params["refresh"] != "")
  }
}

func profileHandler(ctx *web.Context, userId string) {
  loggedInUser := getLoggedInUser(ctx)
  var user User
  if len(userId) == 0 {
    user = loggedInUser
  } else {
    user = GetUserFromUserName(userId, getNotifier(ctx))
  }
  addedQuestions := getQuestionsFromId(user.AddedQuestionIds, getNotifier(ctx))
  Render("profile", profileData{templateData:templateData{User: loggedInUser, Context: ctx}, Profile: user, AddedQuestions: addedQuestions}, ctx, ctx.Params["refresh"] != "")
}

func editQuestionHandlerGen(save bool) func(*web.Context, string) {
  return func(ctx *web.Context, questionId string) {
    user := getLoggedInUser(ctx)
    question := Question{}
    if len(questionId) > 0 {
      question = getQuestionFromId(questionId, getNotifier(ctx))
      if question.AddedUserId != user.UserId {
        getNotifier(ctx)("danger", "Question was added by differentuser, you can't edit it.")
        Render("empty", templateData{User: user, Context: ctx}, ctx, ctx.Params["refresh"] != "")
        return
      }
      if save {
        options, correctoption := getOptions(ctx.Params["options"], ctx.Params["correctoptionindex"], getNotifier(ctx))
        question.Question = ctx.Params["question"]
        question.Options = options
        question.CorrectOption = correctoption
        question.Save()
        getNotifier(ctx)("success", "Question saved successfully")
      }
    }
    Render("editquestion", editQuestionData{templateData:templateData{User: user, Context: ctx}, Question: question}, ctx, ctx.Params["refresh"] != "")
  }
}

func deleteQuestionHandler(ctx *web.Context, questionId string) {
  deleteQuestion(questionId, getNotifier(ctx))
  ctx.Redirect(301, "/myprofile/")
}

func questionHandler(ctx *web.Context, questionId string) {
  user := getLoggedInUser(ctx)
  question := Question{}
  if len(questionId) > 0 {
    question = getQuestionFromId(questionId, getNotifier(ctx))
  } else {
    question = getRandomQuestion()
  }
  Render("question", questionData{templateData:templateData{User: user, Context: ctx}, Question:question}, ctx, ctx.Params["refresh"] != "")
}

func loginSubmitHandler(ctx *web.Context) {
  user := LogIn(ctx.Params["username"], ctx.Params["password"], getNotifier(ctx))
  if len(user.UserId) == 0 {
    simplePageHandler("login")(ctx)
    return
  }
  setSecureCookie(ctx, "userid", user.UserId, 0)
  ctx.Redirect(301, "/home/")
}

func logoutHandler(ctx *web.Context) {
  setSecureCookie(ctx, "userid", "", -1) // Deleting cookie
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
  user = SignUp(user, getNotifier(ctx))
  if len(user.UserId) == 0 {
    simplePageHandler("signup")(ctx)
    return
  }
  setCookie(ctx, "userid", user.UserId, 0)
  ctx.Redirect(301, "/home/")
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

func addQuestionSubmitHandler(ctx *web.Context) {
  loggedInUser := getLoggedInUser(ctx)
  options, correctoption := getOptions(ctx.Params["options"], ctx.Params["correctoptionindex"], getNotifier(ctx))
  if len(options) > 0 {
    question := Question {
      QuestionId: GetNextId("question"),
      Question: ctx.Params["question"],
      Options: options,
      CorrectOption: correctoption,
      AddedUserId:loggedInUser.UserId,
    }
    AddQuestion(question, getNotifier(ctx))
    ctx.Redirect(301, "/question/" + question.QuestionId + "/edit/")
  }
  simplePageHandler("editquestion")(ctx)
}

