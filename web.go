package main

import (
	"fmt"
	"net/http"
	"os"
  "os/signal"
  "./renderer"
  "./log"
  "./datastore"
  "github.com/gorilla/context"
)

type homeData struct {
  User datastore.User
}

type questionData struct {
  User datastore.User
  Question datastore.Question
}

type loginData struct {
  User datastore.User
  Info string
}

func main() {

	bind := fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))

  // Handlers
  http.HandleFunc("/", homeHandler)
	http.HandleFunc("/quiz/", quizHandler)
  http.HandleFunc("/login/", loginHandler)

  // Setting up hooks to close db connections
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt)
  go func() {
    for sig := range c {
      fmt.Printf("%+v\n", sig)
      datastore.CloseOpenSessions()
      os.Exit(1)
    }
  }()

  // Starting webserver
  log.Info(fmt.Sprintf("listening on %s...", bind))
	err := http.ListenAndServe(bind, context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		panic(err)
	}
}

func homeHandler(res http.ResponseWriter, req *http.Request) {
  session := datastore.GetSession("sname", req)
  user := datastore.GetUserFromId(session.GetLoggedInUserId())
  renderer.Render("home", homeData{User: user}, res, req)
  session.Save(res, req)
}

func loginHandler(res http.ResponseWriter, req *http.Request) {
  session := datastore.GetSession("sname", req)
  loginInfo := ""
  if req.PostFormValue("username") != "" {
    user := datastore.User{}
    user, loginInfo = datastore.LogIn(req.PostFormValue("username"), req.PostFormValue("password"))
    session.SetLoggedInUser(user)
  }
  loggedInUser := datastore.GetUserFromId(session.GetLoggedInUserId())
  renderer.Render("login", loginData{User: loggedInUser, Info: loginInfo}, res, req)
  session.Save(res, req)
}

func addQuestionHandler(res http.ResponseWriter, req *http.Request) {
  err := datastore.AddQuestion(datastore.Question{Question:"How do you move to right?", Options:[]string{"right command", "r key", "left command", "l key"}})

  if err != nil {
    panic(err)
  }

  fmt.Fprintf(res, "Done")
}

func quizHandler(res http.ResponseWriter, req *http.Request) {
  session := datastore.GetSession("sname", req)
  user := datastore.GetUserFromId(session.GetLoggedInUserId())
  renderer.Render("question", questionData{User: user, Question: datastore.GetRandomQuestion()}, res, req)
  session.Save(res, req)
}
