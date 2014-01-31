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

func GetUserFromBson(query bson.M) (User, []alert) {
  users := []User{}
  err := uCollection.Find(query).All(&users)
  if err != nil {
    panic(err)
  }
  return getUserOrError(users)
}

func GetUserFromUserName(username string) (User, []alert) {
  return GetUserFromBson(bson.M{"username":username})
}

func GetUserFromId(userid string) (User, []alert) {
  return GetUserFromBson(bson.M{"userid":userid})
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

func SignUp(user User) (User, []alert) {
  err := uCollection.Insert(&user)
  if err != nil {
    panic(err)
    return user, []alert{alert{Text:"Internal error in creating user, try again later", Type:"danger"}}
  }
  return user, []alert{}
}

func GetUsersCollection(db string) *mgo.Collection {
  return GetCollection(GetMgoConnection(), db, "users")
}
