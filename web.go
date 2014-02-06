package main

import (
	"os"
  "os/signal"
  "github.com/hoisie/web"
)

/*
   TODOs:
   * Profile pic
   * Send verification mail again
   * Edit profile
   * Home page flowchart
   * server validation
   * implement returnto for request that will redirect
   * Discussion on question
   * Collect stats for questions and users => and grade question and users to show question of relevant difficulty to every user
*/

func main() {

  // Configs
  web.Config.CookieSecret = getenv("COOKIESECRET")

  // Handlers
  web.Get("/", func(ctx *web.Context) { ctx.Redirect(301, "/home/") })
  web.Get("/home/", simplePageHandler("home"))
  web.Get("/login/", simplePageHandler("login", func(ctx *web.Context) bool {
    user := getLoggedInUser(ctx)
    if user.UserId != "" {
      getNotifier(ctx)("info", "User already logged in as " + user.DisplayName)
      ctx.Redirect(301, "/home/")
      return true
    }
    return false
  }))
  web.Get("/signup/", simplePageHandler("signup"))
  web.Get("/addquestion/()", editQuestionHandlerGen(false))
  web.Get("/myprofile/()", profileHandler)
  web.Get("/user/(.*)/", profileHandler)
  web.Get("/question/(.*)/edit/",editQuestionHandlerGen(false))
  web.Get("/question/(.*)/", questionHandler)
  web.Get("/emailverification/(.*)/(.*)", verificationHandler)
  web.Get("/checkusernameavailability/(.*)", availabilityHandler)

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
      logger.Println(sig)
      CloseOpenSessions()
      os.Exit(1)
    }
  }()

  // Starting webserver
  web.Run(getHostPort())
}

func getBaseUrl() string {
  return "http://" + getHostPort()
}

func getHostPort() string {
  if getenv("PORT") == "80" {
    return getenv("HOST")
  }
  return getenv("HOST") + ":" + getenv("PORT")
}
