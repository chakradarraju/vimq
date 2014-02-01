package main

import (
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
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

func AddQuestion(question Question, notifier func(string,string)) {
  err := qCollection.Insert(&question)
  if err == nil {
    user := GetUserFromId(question.AddedUserId, notifier)
    user.NewAddedQuestionId(question.QuestionId)
    user.Save()
    notifier("info", "Question created successfully")
    return
  }
  panic(err)
}

func getQuestionsFromId(questionIds []string, notifier func(string,string)) []Question {
  qdata := make([]Question, len(questionIds))
  for i, questionId := range questionIds {
    qdata[i] = getQuestionFromId(questionId, notifier)
  }
  return qdata
}

func getQuestionFromId(id string, notifier func(string,string)) Question {
  questions := []Question{}
  if err := qCollection.Find(bson.M{"questionid": id}).All(&questions); err != nil {
    panic(err)
  }
  if len(questions) < 1 {
    notifier("danger", "Question with id " + id + " not found")
    return Question{}
  } else if len(questions) > 1 {
    notifier("warning", "More than one question with id " + id + " found")
  }
  return questions[0]
}

func GetQuestionsCollection(db string) *mgo.Collection {
  return GetCollection(GetMgoConnection(), db, "questions")
}

func (q *Question) Save() {
  qCollection.Update(bson.M{"questionid":q.QuestionId}, &q)
}
