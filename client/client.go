package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "time"

    "github.com/goperfapps/microservices/catalog"
    "google.golang.org/grpc"
    //"google.golang.org/grpc/credentials/insecure"
)

type SignupRequest struct {
    FirstName string `json:"firstName"`
    LastName  string `json:"lastName"`
    Email     string `json:"email"`
    Password  string `json:"password"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type AuthResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

func makeSignupRequest(url string, req SignupRequest) (*AuthResponse, error) {
    reqJson, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqJson))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    log.Printf("Signup response body: %s\n", string(body))

    var authRes AuthResponse
    if err := json.Unmarshal(body, &authRes); err != nil {
        return &AuthResponse{
            Success: false,
            Message: string(body),
        }, nil
    }

    return &authRes, nil
}

func makeLoginRequest(url string, req LoginRequest) (*AuthResponse, error) {
    reqJson, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqJson))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    log.Printf("Login response body: %s\n", string(body))

    var authRes AuthResponse
    if err := json.Unmarshal(body, &authRes); err != nil {
        return &AuthResponse{
            Success: false,
            Message: string(body),
        }, nil
    }

    return &authRes, nil
}

func getProductByID(client catalog.CatalogServiceClient, productID int32) (*catalog.Product, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req := &catalog.GetProductByIdRequest{Id: productID}
    res, err := client.GetProductById(ctx, req)
    if err != nil {
        return nil, err
    }

    return res, nil
}

func main() {
    // Sign up request
    signupReq := SignupRequest{
        FirstName: "John",
        LastName:  "Doe",
        Email:     "john.doe@example.com",
        Password:  "password123",
    }
    signupURL := "http://localhost:50053/signup"
    signupRes, err := makeSignupRequest(signupURL, signupReq)
    if err != nil {
        log.Fatalf("Signup request failed: %v", err)
    }
    fmt.Printf("Signup response: %+v\n", signupRes)

    // Login request
    loginReq := LoginRequest{
        Email:    "john.doe@example.com",
        Password: "password123",
    }
    loginURL := "http://localhost:50053/login"
    loginRes, err := makeLoginRequest(loginURL, loginReq)
    if err != nil {
        log.Fatalf("Login request failed: %v", err)
    }
    fmt.Printf("Login response: %+v\n", loginRes)

    // Connect to catalog server
    conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Failed to connect to catalog server: %v", err)
    }
    defer conn.Close()
    catalogClient := catalog.NewCatalogServiceClient(conn)

    // Get product by ID
    productID := int32(1)
    product, err := getProductByID(catalogClient, productID)
    if err != nil {
        log.Fatalf("Failed to get product: %v", err)
    }
    fmt.Printf("Product: %+v\n", product)
}

