package main

import (
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "strconv"
)

var (
  openSessions = map[string]*mgo.Session{}
  sCollection = GetCollection(GetConnection(getenv("OPENSHIFT_MONGODB_DB_HOST")+":"+getenv("OPENSHIFT_MONGODB_DB_PORT")), getenv("DB"), "sequence")
)

type Sequence struct {
  Counter string
  NextId int
}

func GetConnection(server string) *mgo.Session {
  if openSessions[server] != nil {
    logger.Println("Reusing connection to " + server)
    return openSessions[server]
  }
  logger.Println("Opening a new connection to " + server + "...")
  session, err := mgo.Dial(server)
  if err != nil {
    panic(err)
  }
  session.SetMode(mgo.Monotonic, true)
  
  openSessions[server] = session
  return session
}

func GetCollection(session *mgo.Session, db string, collection string) *mgo.Collection {
  return session.DB(db).C(collection)
}

func GetNextId(collection string) string {
  sequence := Sequence{}
  change := mgo.Change {
    Update: bson.M{"$inc": bson.M{"NextId": 1}},
    ReturnNew: true,
  }
  _, err := sCollection.Find(bson.M{"Counter": collection}).Apply(change, &sequence)
  if err != nil {
    panic(err)
  }
  return strconv.Itoa(sequence.NextId)
}

func CloseOpenSessions() {
  logger.Println("Closing open sessions")
  for _, session := range openSessions {
    session.Close()
  }
}
