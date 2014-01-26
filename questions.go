package main

import (
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
)

var (
  qCollection *mgo.Collection = GetQuestionsCollection("localhost", "vquiz")
)

type Question struct {
  QuestionId string
  Question string
  Options []string
  CorrectOption string
}

func GetRandomQuestion() Question {
  qdata := Question{}
  err := qCollection.Find(bson.M{}).One(&qdata)
  if err != nil {
    panic(err)
  }
  return qdata
}

func AddQuestion(question Question) []alert {
  err := qCollection.Insert(&question)
  if err == nil {
    return []alert{alert{Text:"Question added successfully", Type:"info"}}
  }
  panic(err)
  return []alert{alert{Text:"Error in adding question", Type:"error"}}
}

func GetQuestionsCollection(server string, db string) *mgo.Collection {
  return GetCollection(GetConnection(server), db, "questions")
}
