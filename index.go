package main

import (
//    "crypto/tls"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
)

const (
	DBHost  =
	DBPort  =
	DBUser  =
	DBPass  =
	DBDbase =
	PORT    =
)

var database *sql.DB

type Page struct {
	Title   string
	Content template.HTML
	Date    string
    GUID    string
}
//Truncate p.Content if execeeds 150 chars
func (p Page) TruncatedText() template.HTML {
    chars := 0
    for i, _ := range p.Content {
        chars++
        if chars > 150 {
            return p.Content[:i] + `...`
        }
    }
    return p.Content
}
func (p Page) Count() int {
    return len(p.Content)
}

func ServeTitle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageGUID := vars["guid"]
	thisPage := Page{}
	fmt.Println(pageGUID)

	err := database.QueryRow("SELECT page_title,page_content, page_date FROM pages WHERE page_guid=?", pageGUID).Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println("Couldn't get page!")
		return
	}
	// html := `<html><head><title>` + thisPage.Title + `</title></head><body><h1>` + thisPage.Title + `</h1><div>` + thisPage.Content + `</div></body></html>`

	t, _ := template.ParseFiles("templates/title.html")
	t.Execute(w, thisPage)
}
func ServePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageGUID := vars["guid"]
	thisPage := Page{}
	fmt.Println(pageGUID)
	err := database.QueryRow("SELECT page_title,page_content,page_date FROM pages WHERE page_guid=?", pageGUID).Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println("Couldn't get page!")
		return
	}
	// html := `<html><head><title>` + thisPage.Title + `</title></head><body><h1>` + thisPage.Title + `</h1><div>` + thisPage.Content + `</div></body></html>`

	t, _ := template.ParseFiles("templates/blog.html")
	t.Execute(w, thisPage)
}
func RedirIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home", 301)
}
func ServeIndex(w http.ResponseWriter, r *http.Request) {
    var Pages = []Page{}
	pages, err := database.Query("SELECT page_title,page_content,page_date,page_guid FROM pages ORDER BY ? DESC", "page_date")
	if err != nil {
		fmt.Fprintln(w, err.Error)
	}
    for pages.Next() {
		thisPage := Page{}
		pages.Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date, &thisPage.GUID)
		Pages = append(Pages, thisPage)
	}
    t , _ := template.ParseFiles("templates/index.html")
    t.Execute(w, Pages)

}


func main() {
	dbConn := fmt.Sprintf("%s:%s@/%s", DBUser, DBPass, DBDbase)
	fmt.Println(dbConn)
	db, err := sql.Open("mysql", dbConn)
	if err != nil {
		log.Println("Couldn't connect!")
		log.Println(err.Error)
	}
	database = db

	routes := mux.NewRouter()
	routes.HandleFunc("/title/{guid:[0-9a-zA\\-]+}", ServeTitle)
	routes.HandleFunc("/page/{guid:[0-9a-zA\\-]+}", ServePage)
    routes.HandleFunc("/home" , ServeIndex)
    routes.HandleFunc("/", RedirIndex)
	http.Handle("/", routes)

	http.ListenAndServe(PORT, nil)
}
