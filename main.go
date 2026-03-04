package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

type Series struct {
	ID             int
	Name           string
	CurrentEpisode int
	TotalEpisodes  int
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./series.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/decrement", decrementHandler)

	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS series (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		current_episode INTEGER NOT NULL,
		total_episodes INTEGER NOT NULL
	);`
	db.Exec(query)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, current_episode, total_episodes FROM series")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var seriesList []Series

	for rows.Next() {
		var s Series
		rows.Scan(&s.ID, &s.Name, &s.CurrentEpisode, &s.TotalEpisodes)
		seriesList = append(seriesList, s)
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, seriesList)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("templates/create.html"))
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()

		name := r.FormValue("series_name")
		currentEp := r.FormValue("current_episode")
		totalEps := r.FormValue("total_episodes")

		currentInt, _ := strconv.Atoi(currentEp)
		totalInt, _ := strconv.Atoi(totalEps)

		if name == "" || currentInt < 1 || totalInt < 1 || currentInt > totalInt {
			http.Error(w, "Invalid input", 400)
			return
		}

		_, err := db.Exec(
			"INSERT INTO series (name, current_episode, total_episodes) VALUES (?, ?, ?)",
			name, currentInt, totalInt,
		)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	id := r.URL.Query().Get("id")

	_, err := db.Exec(`
		UPDATE series
		SET current_episode = current_episode + 1
		WHERE id = ? AND current_episode < total_episodes
	`, id)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("ok"))
}

func decrementHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	id := r.URL.Query().Get("id")

	_, err := db.Exec(`
		UPDATE series
		SET current_episode = current_episode - 1
		WHERE id = ? AND current_episode > 1
	`, id)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("ok"))
}
