package main

import (
  "github.com/hoisie/web"
)

type data map[string]interface{}

func availabilityHandler(ctx *web.Context, username string) []byte {
  ret, _ := encodeJson(data {
    "username": username,
    "availability": checkUserNameAvailability(username, getNotifier(ctx)),
    "alerts": getNotifications(ctx),
  })
  return ret
}

func verificationHandler(ctx *web.Context, userId string, hash string) {
  verifyUser(userId, hash, getNotifier(ctx))
  Render("mailverified", data {
    "Context": ctx,
    "Alerts": getNotifications(ctx),
  }, ctx, ctx.Params["refresh"] != "")
}

func simplePageHandler(page string, modifiers ...func(*web.Context) bool) func(*web.Context) {
  return func(ctx *web.Context) {
    for _, fn := range modifiers {
      if fn(ctx) {
        return
      }
    }
    user := getLoggedInUser(ctx)
    Render(page, data {
      "User": user,
      "Context": ctx,
      "Alerts": getNotifications(ctx),
    }, ctx, ctx.Params["refresh"] != "")
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
  Render("profile", data {
    "User": loggedInUser, 
    "Context": ctx,
    "Alerts": getNotifications(ctx),
    "Profile": user,
    "AddedQuestions": addedQuestions,
  }, ctx, ctx.Params["refresh"] != "")
}

func editQuestionHandlerGen(save bool) func(*web.Context, string) {
  return func(ctx *web.Context, questionId string) {
    user := getLoggedInUser(ctx)
    question := Question{}
    if len(questionId) > 0 {
      question = getQuestionFromId(questionId, getNotifier(ctx))
      if question.AddedUserId != user.UserId {
        getNotifier(ctx)("danger", "Question was added by differentuser, you can't edit it.")
        Render("empty", data {
          "User": user,
          "Context": ctx,
          "Alerts": getNotifications(ctx),
        }, ctx, ctx.Params["refresh"] != "")
        return
      }
      if save {
        options, correctoption := getOptions(ctx.Params["options"], ctx.Params["correctoptionindex"], getNotifier(ctx))
        question.Question = ctx.Params["question"]
        question.Options = options
        question.CorrectOption = correctoption
        question.Explanation = ctx.Params["explanation"]
        question.Save()
        getNotifier(ctx)("success", "Question saved successfully")
      }
    }
    Render("editquestion", data {
      "User": user,
      "Context": ctx,
      "Alerts": getNotifications(ctx),
      "Question": question,
    }, ctx, ctx.Params["refresh"] != "")
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
  Render("question", data {
    "User": user,
    "Context": ctx,
    "Alerts": getNotifications(ctx),
    "Question": question,
  }, ctx, ctx.Params["refresh"] != "")
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
  user := getLoggedInUser(ctx)
  if user.UserId == "" {
    getNotifier(ctx)("danger", "User not logged in to logout")
  }
  setSecureCookie(ctx, "userid", "", -1) // Deleting cookie
  ctx.Redirect(301, "/home/")
}

func signupSubmitHandler(ctx *web.Context) {
  if !checkUserNameAvailability(ctx.Params["username"], getNotifier(ctx)) {
    getNotifier(ctx)("info", "Username already registered")
    simplePageHandler("signup")(ctx)
    return
  }
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
  ctx.Redirect(301, "/home/")
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
      Explanation: ctx.Params["explanation"],
    }
    AddQuestion(question, getNotifier(ctx))
    ctx.Redirect(301, "/question/" + question.QuestionId + "/edit/")
    return
  }
  simplePageHandler("editquestion")(ctx)
}
