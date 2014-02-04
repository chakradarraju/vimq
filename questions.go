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
  Explanation string
}

func getRandomQuestion() Question {
  qdata := Question{}
  err := qCollection.Find(bson.M{}).One(&qdata)
  if err != nil {
    panic(err)
  }
  return qdata
}

func AddQuestion(question Question, notify func(string,string)) {
  err := qCollection.Insert(&question)
  if err == nil {
    user := GetUserFromId(question.AddedUserId, notify)
    user.NewAddedQuestionId(question.QuestionId)
    user.Save()
    logger.Println("Added notification")
    notify("info", "Question created successfully")
    return
  }
  panic(err)
}

func getQuestionsFromId(questionIds []string, notify func(string,string)) []Question {
  qdata := make([]Question, len(questionIds))
  for i, questionId := range questionIds {
    qdata[i] = getQuestionFromId(questionId, notify)
  }
  return qdata
}

func getQuestionFromId(id string, notify func(string,string)) Question {
  questions := []Question{}
  if err := qCollection.Find(bson.M{"questionid": id}).All(&questions); err != nil {
    panic(err)
  }
  if len(questions) < 1 {
    notify("danger", "Question with id " + id + " not found")
    return Question{}
  } else if len(questions) > 1 {
    notify("warning", "More than one question with id " + id + " found")
  }
  return questions[0]
}

func GetQuestionsCollection(db string) *mgo.Collection {
  return GetCollection(GetMgoConnection(), db, "questions")
}

func (q *Question) Save() {
  qCollection.Update(bson.M{"questionid":q.QuestionId}, &q)
}

func deleteQuestion(id string, notify func(string,string)) {
  question := getQuestionFromId(id, notify)
  if err := qCollection.Remove(bson.M{"questionid":id}); err == nil {
    user := GetUserFromId(question.AddedUserId, notify)
    user.RemoveQuestionId(id)
    user.Save()
    notify("info", "Question removed successfully")
  } else {
    notify("danger", "Error in removing question, try again later")
  }
}
