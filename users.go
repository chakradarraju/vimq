package main

import (
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
)

type User struct {
  UserId string
  UserName string
  PassWord string
  UserLevel string
  EmailId string
  DisplayName string
  AddedQuestionIds []string
}

var (
  uCollection *mgo.Collection = GetUsersCollection(getenv("DB"))
)

func (u *User) NewAddedQuestionId(questionId string) {
  u.AddedQuestionIds = append(u.AddedQuestionIds, questionId)
}

func (u *User) Save() {
  uCollection.Update(bson.M{"userid":u.UserId}, &u)
}

func GetUserFromBson(query bson.M, notifier func(string,string)) User {
  users := []User{}
  err := uCollection.Find(query).All(&users)
  if err != nil {
    panic(err)
  }
  return getUserOrError(users, notifier)
}

func GetUserFromUserName(username string, notifier func(string,string)) User {
  return GetUserFromBson(bson.M{"username":username}, notifier)
}

func GetUserFromId(userid string, notifier func(string,string)) User {
  return GetUserFromBson(bson.M{"userid":userid}, notifier)
}

func getUserOrError(users []User, notifier func(string,string)) User {
  if len(users) == 0 {
    notifier("danger", "User not found")
    return User{}
  }
  if len(users) > 1 {
    notifier("danger", "More than one user found")
    return User{}
  }
  return users[0]
}

func LogIn(username string, password string, notifier func(string,string)) User {
  user := GetUser(username, notifier)
  if len(user.UserId) == 0 {
    return user
  }
  if user.PassWord == password {
    return user
  }
  notifier("danger", "User password incorrect")
  return User{}
}

func GetUser(username string, notifier func(string,string)) User {
  users := []User{}
  err := uCollection.Find(bson.M{"username": username}).All(&users)
  if err != nil {
    panic(err)
  }
  user := getUserOrError(users, notifier)
  if len(user.UserId) == 0 {
    return user
  }
  return user
}

func SignUp(user User, notifier func(string,string)) User {
  err := uCollection.Insert(&user)
  if err != nil {
    panic(err)
    notifier("danger", "Internal error in creating user, try again later")
    return user
  }
  return user
}

func GetUsersCollection(db string) *mgo.Collection {
  return GetCollection(GetMgoConnection(), db, "users")
}
