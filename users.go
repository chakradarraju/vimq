package main

import (
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "crypto/hmac"
  "crypto/sha1"
  "fmt"
  "regexp"
)

type User struct {
  UserId string
  UserName string
  PassWord string
  UserLevel string
  EmailId string
  EmailVerified string
  DisplayName string
  AddedQuestionIds []string
}

var (
  uCollection *mgo.Collection = GetUsersCollection(getenv("DB"))
)

func (u *User) SetEmail(email string, notify func(string,string)) bool {
  if match, _ := regexp.MatchString("[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}", email); match {
    u.EmailId = email
    sendVerificationMail(*u)
    return true
  }
  notify("warning", "Please check your email id, we suspect it is invalid.");
  return false
}

func (u *User) SetDisplayName(displayname string, notify func(string,string)) bool {
  if match, _ := regexp.MatchString("^.{6,64}$", displayname); match {
    u.DisplayName = displayname
    return true
  }
  notify("warning", "Display name should be at least 6, at most 64 characters")
  return false
}

func (u *User) RemoveQuestionId(questionId string) {
  pos := -1
  for i, id := range u.AddedQuestionIds {
    if id == questionId {
      pos = i
    }
  }
  if pos != -1 {
    l := len(u.AddedQuestionIds)
    u.AddedQuestionIds[pos] = u.AddedQuestionIds[l-1]
    u.AddedQuestionIds = u.AddedQuestionIds[0:l-1]
  }
}

func (u *User) NewAddedQuestionId(questionId string) {
  u.AddedQuestionIds = append(u.AddedQuestionIds, questionId)
}

func (u *User) Save() {
  uCollection.Update(bson.M{"userid":u.UserId}, &u)
}

func getUserFromBson(query bson.M, notify func(string,string)) User {
  users := []User{}
  err := uCollection.Find(query).All(&users)
  if err != nil {
    panic(err)
  }
  return getUserOrError(users, notify)
}

func getUserFromUserName(username string, notify func(string,string)) User {
  return getUserFromBson(bson.M{"username":username}, notify)
}

func getUserFromId(userid string, notify func(string,string)) User {
  return getUserFromBson(bson.M{"userid":userid}, notify)
}

func getUserOrError(users []User, notify func(string,string)) User {
  if len(users) == 0 {
    notify("danger", "User not found")
    return User{}
  }
  if len(users) > 1 {
    notify("danger", "More than one user found")
    return User{}
  }
  return users[0]
}

func LogIn(username string, password string, notify func(string,string)) User {
  user := GetUser(username, notify)
  if len(user.UserId) == 0 {
    return user
  }
  if user.PassWord == password {
    return user
  }
  notify("danger", "User password incorrect")
  return User{}
}

func GetUser(username string, notify func(string,string)) User {
  users := []User{}
  err := uCollection.Find(bson.M{"username": username}).All(&users)
  if err != nil {
    panic(err)
  }
  user := getUserOrError(users, notify)
  if len(user.UserId) == 0 {
    return user
  }
  return user
}

func SignUp(user User, notify func(string,string)) User {
  err := uCollection.Insert(&user)
  if err != nil {
    panic(err)
    notify("danger", "Internal error in creating user, try again later")
    return user
  }
  sendVerificationMail(user)
  notify("info", "Verification mail has been sent, please verify your email id")
  return user
}

func checkUserNameAvailability(username string, notify func(string,string)) bool {
  user := getUserFromUserName(username, func(a string, b string) {})
  return user.UserId == ""
}

func sendVerificationMail(user User) {
  if user.EmailId == user.EmailVerified {
    return
  }
  mail([]string{user.EmailId}, renderMail("verification", map[string]interface{} {
    "From": "vimqmail+noreply@gmail.com",
    "To": user.EmailId,
    "Subject": "Email Verification from VimQ",
    "VerificationLink": getBaseUrl() + "/emailverification/" + user.UserId + "/" + getUserHash(user.UserId, user.EmailId),
  }))
}

func getUserHash(userId string, emailId string) string {
  hm := hmac.New(sha1.New, []byte(getenv("COOKIESECRET")))

  hm.Write([]byte(userId))
  hm.Write([]byte(emailId))
  return fmt.Sprintf("%02x", hm.Sum(nil))
}

func verifyUser(userId string, hash string, notify func(string,string)) {
  user := getUserFromId(userId, notify)
  if getUserHash(userId, user.EmailId) == hash {
    user.EmailVerified = user.EmailId
    user.Save()
    notify("info", "Your email id is verified")
  } else {
    notify("danger", "Hash doesn't seem to match")
  }
}

func GetUsersCollection(db string) *mgo.Collection {
  return GetCollection(GetMgoConnection(), db, "users")
}
