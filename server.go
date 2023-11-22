package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

type LoginInfo struct {
	Email    string
	Password string
}

func main() {
	// Connect to the database
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Create the "user" table if it does not exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user (
			email VARCHAR PRIMARY KEY,
			password VARCHAR
		);
	`)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	// Insert a sample user if it does not exist
	existingUser := struct{ Email string }{}
	err = db.QueryRow("SELECT email FROM user WHERE email=?", "higuys@gmail.com").Scan(&existingUser.Email)
	if err == sql.ErrNoRows {
		result, err := db.Exec("INSERT INTO user (email, password) VALUES (?, ?)", "higuys@gmail.com", "hg12345")
		if err != nil {
			log.Fatal("Error inserting value:", err)
		}
		_ = result
	} else if err != nil {
		log.Fatal("Error checking existing user:", err)
	}

	// Create login template and handle form request
	tmpl := template.Must(template.ParseFiles("login.html"))
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			tmpl.Execute(w, nil)
			return
		}

		// Get values from user input
		logInfo := LoginInfo{
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}

		// Compare user input with the database
		var dbPassword string
		tmplResult := struct{ Success, Error bool }{false, false}

		err := db.QueryRow("SELECT password FROM user WHERE email=?", logInfo.Email).Scan(&dbPassword)
		if err != nil {
			tmplResult.Error = true
		} else if logInfo.Password == dbPassword {
			tmplResult.Success = true
		} else {
			tmplResult.Error = true
		}

		tmpl.Execute(w, tmplResult)
	})

	// Create server
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":50", nil))
}
