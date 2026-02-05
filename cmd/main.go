package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"


	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	db    *sql.DB
	store = sessions.NewCookieStore([]byte("super-secret-key"))
)

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./esemos.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		password TEXT
	);
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		slug TEXT UNIQUE,
		content TEXT,
		summary TEXT,
		author TEXT,
		created_at DATETIME,
		updated_at DATETIME
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Create default admin if not exists
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", "admin", string(hashedPassword))
		fmt.Println("Default admin created: admin / admin123")
	}
}

func main() {
	initDB()
	defer db.Close()

	r := mux.NewRouter()

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Public routes
	r.HandleFunc("/", handleHome).Methods("GET")
	r.HandleFunc("/news", handleNews).Methods("GET")
	r.HandleFunc("/news/{slug}", handlePostDetail).Methods("GET")
	r.HandleFunc("/kontakt", handleContact).Methods("GET", "POST")

	// Admin routes
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(authMiddleware)
	admin.HandleFunc("", handleAdminDashboard).Methods("GET")
	admin.HandleFunc("/posts", handleAdminPosts).Methods("GET")
	admin.HandleFunc("/posts/new", handleAdminNewPost).Methods("GET", "POST")
	admin.HandleFunc("/posts/edit/{id}", handleAdminEditPost).Methods("GET", "POST")
	admin.HandleFunc("/posts/delete/{id}", handleAdminDeletePost).Methods("POST")

	// Login/Logout
	r.HandleFunc("/login", handleLogin).Methods("GET", "POST")
	r.HandleFunc("/logout", handleLogout).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// Auth Middleware
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session-name")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Handlers (Stubs for now)
func handleHome(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
}

func handleNews(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "news.html", nil)
}

func handlePostDetail(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "post.html", nil)
}

func handleContact(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "contact.html", nil)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, "login.html", nil)
		return
	}
	// Simple login logic
	username := r.FormValue("username")
	password := r.FormValue("password")

	var hash string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		http.Redirect(w, r, "/login?error=1", http.StatusFound)
		return
	}

	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Save(r, w)
	http.Redirect(w, r, "/admin", http.StatusFound)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin/dashboard.html", nil)
}

func handleAdminPosts(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin/posts.html", nil)
}

func handleAdminNewPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Save post logic
		http.Redirect(w, r, "/admin/posts", http.StatusFound)
		return
	}
	renderTemplate(w, "admin/post_form.html", nil)
}

func handleAdminEditPost(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin/post_form.html", nil)
}

func handleAdminDeletePost(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/posts", http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	layoutPath := "web/templates/layout.html"
	tmplPath := "web/templates/" + tmpl

	// Prüfen ob Dateien existieren
	if _, err := os.Stat(layoutPath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Layout template not found at %s. Please ensure you are running the app from the project root.", layoutPath), http.StatusInternalServerError)
		return
	}
	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Content template not found at %s", tmplPath), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles(layoutPath, tmplPath)
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		http.Error(w, fmt.Sprintf("Error parsing templates: %v", err), http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
	}
}
