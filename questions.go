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
  CorrectOptionIndex int
}

func GetRandomQuestion() Question {
  qdata := Question{}
  err := qCollection.Find(bson.M{}).One(&qdata)
  if err != nil {
    panic(err)
  }
  return qdata
}

func AddQuestion(question Question) (error, string) {
  err := qCollection.Insert(&question)
  if err == nil {
    return err, "Question added successfully"
  }
  panic(err)
  return err, "Error in adding question, check Log"
}

func GetQuestionsCollection(server string, db string) *mgo.Collection {
  return GetCollection(GetConnection(server), db, "questions")
}
