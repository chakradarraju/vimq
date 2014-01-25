package main

import (
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "fmt"
)

type User struct {
  UserId string
  UserName string
  PassWord string
  UserLevel string
}

var (
  uCollection *mgo.Collection = GetUsersCollection("localhost", "vquiz")
)

func GetUserFromId(userid string) (User, string) {
  user := User{}
  if userid != "" {
    users := []User{}
    err := uCollection.Find(bson.M{"userid": userid}).All(&users)
    if err != nil {
      panic(err)
    }
    user, info := getUserOrError(users)
    if info == "" {
      return user, info
    }
  }
  return user, ""
}

func getUserOrError(users []User) (User, string) {
  if len(users) == 0 {
    return User{}, "User not found"
  }
  if len(users) > 1 {
    return User{}, "More than one user found"
  }
  return users[0], ""
}

func LogIn(username string, password string) (User, string) {
  users := []User{}
  err := uCollection.Find(bson.M{"username": username}).All(&users)
  logger.Println("Username: " + username + ", password: " + password)
  logger.Println(fmt.Sprintf("%+v",users))
  if err != nil {
    panic(err)
  }
  user, info := getUserOrError(users)
  if info != "" {
    return user, info
  }
  if user.PassWord == password {
    return user, ""
  }
  return User{}, "User password incorrect" + user.PassWord + "2" + password + "1" + user.UserName + "3" + user.UserId
}

func SignUp(username string, password string) error {
  user := User{UserId: GetNextId("user"), UserName: username, PassWord: password}
  return uCollection.Insert(&user)
}

func GetUsersCollection(server string, db string) *mgo.Collection {
  return GetCollection(GetConnection(server), db, "users")
}
