package main

import (
  "html/template"
  //"io/ioutil"
  "net/http"
)

type Page struct {
  Title string
  Body string
}

type TmplPage struct {
  Title string
  Body template.HTML
}

func convertPage (p *Page) *TmplPage {
  return &TmplPage{Title: p.Title, Body: template.HTML(p.Body)}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
  tp := convertPage(&Page{Title: "TestPage", Body: `<a href="/">Home</a>`})
  t, _ := template.ParseFiles("main.html")
  t.Execute(w, tp)
}

func main() {
  http.HandleFunc("/", viewHandler)
  http.ListenAndServe(":8080", nil)
}
