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

func getQuestionsFromId(questionIds []string) ([]Question, []alert) {
  qdata := make([]Question, len(questionIds))
  alerts := []alert{}
  for i, questionId := range questionIds {
    newAlerts := []alert{}
    qdata[i], newAlerts = getQuestionFromId(questionId)
    alerts = append(alerts, newAlerts...)
  }
  return qdata, alerts
}

func getQuestionFromId(id string) (Question, []alert) {
  questions := []Question{}
  if err := qCollection.Find(bson.M{"questionid": id}).All(&questions); err != nil {
    panic(err)
  }
  alerts := []alert{}
  if len(questions) < 1 {
    return Question{}, []alert{alert{Text:"Question with id " + id + " not found", Type: "danger"}}
  } else if len(questions) > 1 {
    alerts = []alert{alert{Text:"More than one question with id " + id + " found", Type: "warning"}}
  }
  return questions[0], alerts
}

func GetQuestionsCollection(db string) *mgo.Collection {
  return GetCollection(GetMgoConnection(), db, "questions")
}

func (q *Question) Save() {
  qCollection.Update(bson.M{"questionid":q.QuestionId}, &q)
}
