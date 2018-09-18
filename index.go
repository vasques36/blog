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
    "os"
    "encoding/json"
)
const (
	DBHost  = "127.0.0.1"
    DBPort  = ":3306"
	DBUser  = "root"
	DBDbase = "cms"
    PORT    = ":8080"
)

var database *sql.DB

type Comment struct {
    Id      int
    Name    string
    Email   string
    CommentText string
}

type Page struct {
    Id       int
	Title   string
	Content template.HTML
    RawContent string
	Date    string
    GUID    string
    Comments []Comment
}
type JSONResponse struct {
    Fields map[string]string
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


func ServePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageGUID := vars["guid"]
	thisPage := Page{}
	fmt.Println(pageGUID)
	err := database.QueryRow("SELECT page_title,page_content,page_date,page_guid FROM pages WHERE page_guid=?", pageGUID).
    Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date,&thisPage.GUID)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println("Couldn't get page!")
		return
	}
    comments, err := database.Query("SELECT id, comment_name as Name,comment_email, comment_text FROM comments WHERE page_id=?", thisPage.Id)
    if err != nil {
    log.Println(err)
    }
    for comments.Next() {
    var comment Comment
    comments.Scan(&comment.Id, &comment.Name, &comment.Email,
    &comment.CommentText)
    thisPage.Comments = append(thisPage.Comments, comment)
}

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
func APIPage(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    pageGUID := vars["guid"]
    thisPage := Page{}
    fmt.Println(pageGUID)
    err := database.QueryRow("SELECT page_title,page_content,page_date FROM pages WHERE page_guid=?", pageGUID).
    Scan(&thisPage.Title, &thisPage.RawContent, &thisPage.Date)
    thisPage.Content = template.HTML(thisPage.RawContent)
    if err != nil {
        http.Error(w, http.StatusText(404), http.StatusNotFound)
        log.Println(err)
        return
    }
    APIOutput, _ := json.Marshal(thisPage)
    if err != nil {
        http.Error(w, "", 500)
        return
    }
    fmt.Println(APIOutput)
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintln(w, thisPage)
}

func APICommentPost(w http.ResponseWriter, r *http.Request) {
    var commentAdded string
    err := r.ParseForm()
    if err != nil {
        log.Println(err.Error)
    }
    fmt.Println(r.FormValue)
    name := r.FormValue("name")
    email := r.FormValue("email")
    comments := r.FormValue("comments")

    res, err := database.Exec("INSERT INTO comments SET comment_name=?, comment_email=?, comment_text=?", name, email, comments)

    if err != nil {
        fmt.Println("Not process")
        log.Println(err.Error)
    }

    id, err := res.LastInsertId()
    if err != nil {
        commentAdded = "false"
    } else {
        commentAdded = "true"
    }

    var resp JSONResponse
    resp.Fields["id"] = string(id)
    resp.Fields["added"] = commentAdded
    jsonResp, _ := json.Marshal(resp)
    w.Header().Set("Content-Type", "application/json")
    fmt.Fprintln(w, jsonResp)
}

func main() {

    pass := os.Getenv("DB_PASS")
    dbConn := fmt.Sprintf("%s:%s@/%s", DBUser, pass, DBDbase)
	fmt.Println(dbConn)
	db, err := sql.Open("mysql", dbConn)
	if err != nil {
		log.Println("Couldn't connect!")
		log.Println(err.Error)
	}
	database = db

	routes := mux.NewRouter()
    routes.HandleFunc("/api/pages", APIPage).
        Methods("GET").
        Schemes("http")
    routes.HandleFunc("/api/page/{id:[\\w\\d\\-]+}", APIPage).
        Methods("GET").
        Schemes("http")
    routes.HandleFunc("/api/comments", APICommentPost).
        Methods("POST")
	routes.HandleFunc("/page/{guid:[0-9a-zA\\-]+}", ServePage)
    routes.HandleFunc("/home" , ServeIndex)
    routes.HandleFunc("/", RedirIndex)
	http.Handle("/", routes)
    http.ListenAndServe(PORT, nil)

}




















