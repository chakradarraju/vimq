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

func GetUserFromBson(query bson.M, notify func(string,string)) User {
  users := []User{}
  err := uCollection.Find(query).All(&users)
  if err != nil {
    panic(err)
  }
  return getUserOrError(users, notify)
}

func GetUserFromUserName(username string, notify func(string,string)) User {
  return GetUserFromBson(bson.M{"username":username}, notify)
}

func GetUserFromId(userid string, notify func(string,string)) User {
  return GetUserFromBson(bson.M{"userid":userid}, notify)
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
  return user
}

func GetUsersCollection(db string) *mgo.Collection {
  return GetCollection(GetMgoConnection(), db, "users")
}
