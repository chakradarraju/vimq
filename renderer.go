package main

import (
  "net/http"
  "html/template"
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
  var err error
  t := template.New("base.html")
  myfuncs := template.FuncMap{
    "isTemplateDefined": func(name string) bool {
      return t.Lookup(name) != nil
    },
  }
  t, err = t.Funcs(myfuncs).ParseFiles("templates/base.html", "templates/" + file + ".html", "templates/header.html")
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
