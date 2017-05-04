package main

import (
  "io/ioutil"
  "net/http"
  "html/template"
  "regexp"
  "errors"
)

//var templates = template.Must(template.ParseFiles(tmplDir+"/edit.html", tmplDir+"/view.html"))
var templates = template.Must(template.New("view.html").Funcs(template.FuncMap{"generateLinks": generateLinks,}).ParseFiles(tmplDir+"/edit.html", tmplDir+"/view.html"))
var validPath = regexp.MustCompile("^/(edit|view|save)/([a-zA-Z0-9]+)$")
var pageTitle = regexp.MustCompile(`\[([a-zA-Z0-9]+)\]`)
var pageLink = `<a href="/view/$1">$1</a>`
var dataDir = "data"
var tmplDir = "tmpl"

type Page struct {
  Title string
  Body []byte
}

type TmplPage struct {
  Title string
  Body template.HTML
}

func generateLinks(body []byte) []byte {
  return pageTitle.ReplaceAll(body, []byte(pageLink))
}

func (p *Page) save() error {
  filename := dataDir+"/" + p.Title + ".txt"
  // scan input body before saving
  p.Body = []byte(template.HTMLEscapeString(string(p.Body)))
  return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
  filename := dataDir+"/" + title + ".txt"
  body, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }
  return &Page{Title: title, Body: body}, nil
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
  m := validPath.FindStringSubmatch(r.URL.Path)
  if m == nil {
    http.NotFound(w, r)
    return "", errors.New("Invalid title page.")
  }
  return m[2], nil //the title is the second subexpression
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
  // first convert to TmplPage
  tp := &TmplPage{ Title: p.Title, Body: template.HTML(string(generateLinks(p.Body))) }
  err := templates.ExecuteTemplate(w, tmpl+".html", tp)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
  http.Redirect(w, r, "/view/Home", http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    http.Redirect(w, r, "/edit/"+title, http.StatusFound)
    return
  }
  renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title}
  }
  renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
  body := r.FormValue("body")
  p := &Page{Title: title, Body: []byte(body)}
  err := p.save()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
      http.NotFound(w, r)
      return
    }
    fn(w, r, m[2])
  }
}

func main() {
  fs := http.FileServer(http.Dir("static"))
  http.Handle("/static/", http.StripPrefix("/static/", fs))
  http.HandleFunc("/view/", makeHandler(viewHandler))
  http.HandleFunc("/edit/", makeHandler(editHandler))
  http.HandleFunc("/save/", makeHandler(saveHandler))
  http.HandleFunc("/", indexHandler)
  http.ListenAndServe(":8080", nil)
}
