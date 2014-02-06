package main

import (
  "html/template"
  "github.com/hoisie/web"
  "strings"
  "encoding/json"
)

var (
  cache = map[string]*template.Template{}
)

func loadView(view string) {
  cache[view] = parseTemplate(view)
}

func renderView(ctx *web.Context, template string, data interface{}) {
  if _, found := cache[template]; !found || ctx.Params["refresh"] != "" {
    loadView(template)
  }
  cache[template].Execute(ctx, data)
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
