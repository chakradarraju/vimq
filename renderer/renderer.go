package renderer

import (
  "net/http"
  "html/template"
  "net/url"
  "strings"
  "../log"
)

var (
  parsed map[string]*template.Template = Init([]string{"home","question","login"})
)

func Render(page string, data interface{}, res http.ResponseWriter, req *http.Request) {
  val, err := url.ParseQuery(req.URL.RawQuery)
  if err != nil {
    panic(err)
  }
  if strings.Join(val["refresh"], "") != "" {
    parsed[page] = parseTemplate(page)
  }
  parsed[page].Execute(res, data)
}

func parseTemplate(file string) *template.Template {
  log.Info("Parsing " + file + "...")
  t, err := template.ParseFiles("tmpls/" + file + ".html")
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
