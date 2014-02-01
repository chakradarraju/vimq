package main

import (
  "github.com/hoisie/web"
  "bytes"
  "crypto/hmac"
  "crypto/sha1"
  "encoding/base64"
  "time"
  "strconv"
  "strings"
  "net/http"
  "net/url"
  "fmt"
)

func appendToCookie(ctx *web.Context, name string, value string, age int64) {
  cookie, _ := ctx.Request.Cookie(name)
  if cookie == nil {
    setCookie(ctx, name, url.QueryEscape(value), age)
    return
  }
  setCookie(ctx, name, cookie.Value + url.QueryEscape(";" + value), age)
}

func setCookie(ctx *web.Context, name string, value string, age int64) {
  var expiry time.Time
  if age == 0 {
    expiry = time.Unix(2147483647, 0)
  } else {
    expiry = time.Unix(time.Now().Unix()+age, 0)
  }
  ctx.SetCookie(&http.Cookie{Name: name, Value: value, Expires: expiry, Path: "/"})
}

func setSecureCookie(ctx *web.Context, name string, value string, age int64) {
  if len(ctx.Server.Config.CookieSecret) == 0 {
      ctx.Server.Logger.Println("Secret Key for secure cookies has not been set. Please assign a cookie secret to web.Config.CookieSecret.")
      return
  }
  var buf bytes.Buffer
  encoder := base64.NewEncoder(base64.StdEncoding, &buf)
  encoder.Write([]byte(value))
  encoder.Close()
  vs := buf.String()
  vb := buf.Bytes()
  timestamp := strconv.FormatInt(time.Now().Unix(), 10)
  sig := getCookieSig(ctx.Server.Config.CookieSecret, vb, timestamp)
  cookie := strings.Join([]string{vs, timestamp, sig}, "|")
  setCookie(ctx, name, cookie, age)
}

func getCookieSig(key string, val []byte, timestamp string) string {
  hm := hmac.New(sha1.New, []byte(key))

  hm.Write(val)
  hm.Write([]byte(timestamp))

  hex := fmt.Sprintf("%02x", hm.Sum(nil))
  return hex
}
