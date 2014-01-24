package datastore

import (
  "labix.org/v2/mgo"
  "labix.org/v2/mgo/bson"
  "../log"
)

var (
  openSessions = map[string]*mgo.Session{}
  sCollection = GetCollection(GetConnection("localhost"), "vquiz", "sequence")
)

type Sequence struct {
  Counter string
  NextId int
}

func GetConnection(server string) *mgo.Session {
  if openSessions[server] != nil {
    log.Info("Reusing connection to " + server)
    return openSessions[server]
  }
  log.Info("Opening a new connection to " + server + "...")
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

func GetNextId(collection string) int {
  sequence := Sequence{}
  change := mgo.Change {
    Update: bson.M{"$inc": bson.M{"NextId": 1}},
    ReturnNew: true,
  }
  info, err := sCollection.Find(bson.M{"Counter": collection}).Apply(change, &sequence)
  if err != nil {
    if info != nil {
      log.Info("something")
    }
    panic(err)
  }
  return sequence.NextId
}

func CloseOpenSessions() {
  log.Info("Closing open sessions")
  for _, session := range openSessions {
    session.Close()
  }
}
