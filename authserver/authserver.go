package main

import (
    "encoding/json"
    "log"
    "net/http"
)

type User struct {
    ID       int    `json:"id"`
    FirstName string `json:"firstName"`
    LastName  string `json:"lastName"`
    Email     string `json:"email"`
    Password  string `json:"password"`
}

type AuthResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

// Simulated database to store users
var users []User

func main() {
    // Register HTTP handlers
    http.HandleFunc("/signup", SignupHandler)
    http.HandleFunc("/login", LoginHandler)

    // Start HTTP server
    log.Println("Starting auth server on :50053...")
    if err := http.ListenAndServe(":50053", nil); err != nil {
        log.Fatalf("Failed to start auth server: %v", err)
    }
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
    // Parse request body
    var newUser User
    err := json.NewDecoder(r.Body).Decode(&newUser)
    if err != nil {
        http.Error(w, "Failed to parse request body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    // Simulate signup logic (replace with actual signup logic)
    // For example, check if user already exists
    if userExists(newUser.Email) {
        response := AuthResponse{
            Success: false,
            Message: "User already exists",
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(response)
        return
    }

    // If user does not exist, proceed with signup
    // Replace with actual signup logic to create user in database

    // Simulate adding user to in-memory database
    newUser.ID = len(users) + 1
    users = append(users, newUser)

    // Assuming signup was successful
    response := AuthResponse{
        Success: true,
        Message: "User registered successfully",
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    // Parse request body
    var loginReq User
    err := json.NewDecoder(r.Body).Decode(&loginReq)
    if err != nil {
        http.Error(w, "Failed to parse request body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    // Simulate login logic (replace with actual login logic)
    _, err = loginUser(loginReq.Email, loginReq.Password)
    if err != nil {
        response := AuthResponse{
            Success: false,
            Message: "Invalid email or password",
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(response)
        return
    }

    // If login successful, respond with user information
    response := AuthResponse{
        Success: true,
        Message: "Login successful",
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

// Simulated database operations (replace with actual database interactions)
func userExists(email string) bool {
    for _, user := range users {
        if user.Email == email {
            return true
        }
    }
    return false
}

func loginUser(email, password string) (User, error) {
    for _, user := range users {
        if user.Email == email && user.Password == password {
            return user, nil
        }
    }
    return User{}, nil // Replace with appropriate error handling
}

