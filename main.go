package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	_ "github.com/lib/pq"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func registrationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading request body: %s", err), http.StatusInternalServerError)
		return
	}

	var incomingUser User
	err = json.Unmarshal(body, &incomingUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON data: %s", err), http.StatusBadRequest)
		return
	}

	err = insertUser(incomingUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting user into database: %s", err), http.StatusInternalServerError)
		fmt.Println("Detailed error:", err) // Log the detailed error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Registration successful for user: %s"}`, incomingUser.Username)
}

func insertUser(user User) error {
	const connStr = "user=postgres password=terfo2005 dbname=postgres sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("Error opening database connection: %s", err)
	}
	defer db.Close()

	_, err = db.Exec(`INSERT INTO clients ("user", "password") VALUES ($1, $2)`, user.Username, user.Password)
	if err != nil {
		return fmt.Errorf("Error inserting user into database: %s", err)
	}

	return nil
}

func main() {
	http.HandleFunc("/register", registrationHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("index").Parse(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>User Registration</title>
			</head>
			<body>
				<h2>User Registration</h2>
				<form id="registrationForm" onsubmit="registerUser(event)">
					<label for="username">Username:</label>
					<input type="text" id="username" name="username" required><br>

					<label for="password">Password:</label>
					<input type="password" id="password" name="password" required><br>

					<button type="submit">Register</button>
				</form>
				<script>
					function registerUser(event) {
						event.preventDefault();

						const username = document.getElementById("username").value;
						const password = document.getElementById("password").value;

						const userData = {
							username: username,
							password: password
						};

						fetch("/register", {
							method: "POST",
							headers: {
								"Content-Type": "application/json"
							},
							body: JSON.stringify(userData)
						})
						.then(response => {
							console.log("Response Status:", response.status);
							if (!response.ok) {
								throw new Error("Registration failed");
							}
							return response.json();
						})
						.then(data => {
							console.log("Response Data:", data);
						})
						.catch(error => {
							console.error("Error:", error);
						});
					}
				</script>
			</body>
			</html>
		`)
		if err != nil {
			http.Error(w, "Error rendering HTML", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})

	port := 8080
	fmt.Printf("Server is running on :%d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
