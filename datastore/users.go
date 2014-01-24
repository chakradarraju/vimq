package datastore

import (
  "github.com/gorilla/sessions"
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
)

type User struct {
  UserId int
  UserName string
  PassWord string
}

var (
  uCollection *mgo.Collection = GetUsersCollection("localhost", "vquiz")
)

func GetUserFromId(userid string) User {
  user := User{}
  if userid != "" {
    err := uCollection.Find(bson.M{"UserId": userid}).One(&user)
    if err != nil {
      panic(err)
    }
  }
  return user
}

func LogIn(username string, password string) (User, string) {
  users := []User{}
  err := uCollection.Find(bson.M{"UserName": username}).All(&users)
  if len(users) == 0 {
    return User{}, "User not found in db"
  }
  if len(users) > 1 {
    return User{}, "More than one user with given username"
  }
  if err != nil {
    panic(err)
  }
  if users[0].PassWord == password {
    return users[0], "User verified"
  }
  return User{}, "User password incorrect"
}

func SignUp(session *sessions.Session, username string, password string) error {
  user := User{UserId: GetNextId("user"), UserName: username, PassWord: password}
  return uCollection.Insert(&user)
}

func GetUsersCollection(server string, db string) *mgo.Collection {
  return GetCollection(GetConnection(server), db, "users")
}
