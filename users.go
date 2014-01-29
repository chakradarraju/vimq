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
}

var (
  uCollection *mgo.Collection = GetUsersCollection(getenv("DB"))
)

func GetUserFromId(userid string) (User, []alert) {
  user := User{}
  if userid != "" {
    users := []User{}
    err := uCollection.Find(bson.M{"userid": userid}).All(&users)
    if err != nil {
      panic(err)
    }
    user, info := getUserOrError(users)
    if len(info) == 0 {
      return user, []alert{}
    }
    return user, info
  }
  return user, []alert{}
}

func getUserOrError(users []User) (User, []alert) {
  if len(users) == 0 {
    return User{}, []alert{alert{Text:"User not found", Type:"danger"}}
  }
  if len(users) > 1 {
    return User{}, []alert{alert{Text:"More than one user found", Type:"danger"}}
  }
  return users[0], []alert{}
}

func LogIn(username string, password string) (User, []alert) {
  user, alerts := GetUser(username)
  if len(alerts) > 0 {
    return user, alerts
  }
  if user.PassWord == password {
    return user, []alert{}
  }
  return User{}, []alert{alert{Text:"User password incorrect", Type:"danger"}}
}

func GetUser(username string) (User, []alert) {
  users := []User{}
  err := uCollection.Find(bson.M{"username": username}).All(&users)
  if err != nil {
    panic(err)
  }
  user, error := getUserOrError(users)
  if len(error) > 0 {
    return user, error
  }
  return user, []alert{}
}

func SignUp(username string, password string, email string) (User, []alert) {
user := User{UserId: GetNextId("user"), UserName: username, PassWord: password, EmailId: email, UserLevel: "rookie"}
  err := uCollection.Insert(&user)
  if err != nil {
    panic(err)
  }
  return GetUser(username)
}

func GetUsersCollection(db string) *mgo.Collection {
  return GetCollection(GetMgoConnection(), db, "users")
}
