package main

import (
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "fmt"
)

var (
  qCollection *mgo.Collection = GetQuestionsCollection(getenv("DB"))
)

type Question struct {
  QuestionId string
  Question string
  Options []string
  CorrectOption string
  AddedUserId string
}

func getRandomQuestion() Question {
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
    logger.Println(fmt.Sprintf("%+v",question))
    user, _ := GetUserFromId(question.AddedUserId)
    user.NewAddedQuestionId(question.QuestionId)
    user.Save()
    return []alert{alert{Text:"Question added successfully", Type:"info"}}
  }
  panic(err)
  return []alert{alert{Text:"Error in adding question", Type:"danger"}}
}

func getQuestionsFromId(questionIds []string) []Question {
  qdata := make([]Question, len(questionIds))
  for i, questionId := range questionIds {
    qdata[i] = getQuestionFromId(questionId)
  }
  return qdata
}

func getQuestionFromId(id string) Question {
  question := Question{}
  if err := qCollection.Find(bson.M{"questionid": id}).One(&question); err != nil {
    panic(err)
  }
  return question
}

func GetQuestionsCollection(db string) *mgo.Collection {
  return GetCollection(GetMgoConnection(), db, "questions")
}
