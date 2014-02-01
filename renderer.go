package main

import (
  "net/http"
  "html/template"
  "github.com/hoisie/web"
  "strings"
  "encoding/json"
)

var (
  cache = map[string]*template.Template{}
)

func Render(page string, data interface{}, res http.ResponseWriter, refresh bool) {
  _, found := cache[page]
  if !found || refresh {
    cache[page] = parseTemplate(page)
  }
  cache[page].Execute(res, data)
}

func parseTemplate(file string) *template.Template {
  logger.Println("Parsing " + file + "...")
  myfuncs := template.FuncMap{
    "canAddQuestion": func(user User) bool {
      return user.UserLevel != "" && user.UserLevel != "rookie"
    },
    "addActiveIfNeeded": func(ctx *web.Context, bases ...string) string {
      for _, base := range bases {
        if strings.Index(ctx.Request.URL.Path, base) == 1 {
          return "active"
        }
      }
      return ""
    },
    "marshal": func(v interface {}) template.JS {
      a, _ := json.Marshal(v)
      return template.JS(a)
    },
    "shouldShowEditOptions": func(profile User, loggedIn User) bool {
      return profile.UserId == loggedIn.UserId
    },
  }
  t, err := template.New("base.html").Funcs(myfuncs).ParseFiles("templates/base.html", "templates/" + file + ".html", "templates/header.html")
  if err != nil {
    panic(err)
  }
  return t
}

func Init(pages []string) map[string]*template.Template {
  temp := map[string]*template.Template{}
  for _, page := range pages {
    temp[page] = parseTemplate(page)
  }
  return temp
}
