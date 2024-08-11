package main

import (
    "bytes"
    "context"
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "net/http"
    "strconv"
    "time"

    "github.com/goperfapps/microservices/catalog"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

var (
    httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
        Name: "http_request_duration_seconds",
        Help: "Duration of HTTP requests.",
        Buckets: prometheus.DefBuckets,
    }, []string{"path"})
)

func init() {
    prometheus.MustRegister(httpDuration)
}

type server struct {
    catalog.UnimplementedCatalogServiceServer
    catalogClient catalog.CatalogServiceClient
}

func (s *server) GetProductById(ctx context.Context, req *catalog.GetProductByIdRequest) (*catalog.Product, error) {
    product, err := s.catalogClient.GetProductById(ctx, req)
    if err != nil {
        return nil, err
    }
    return product, nil
}

func hashPassword(password string) string {
    hash := sha256.Sum256([]byte(password))
    return fmt.Sprintf("%x", hash)
}

type AuthResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

func main() {
    conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to catalog server: %v", err)
    }
    defer conn.Close()
    catalogClient := catalog.NewCatalogServiceClient(conn)

    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    s := grpc.NewServer()
    catalog.RegisterCatalogServiceServer(s, &server{catalogClient: catalogClient})
    go func() {
        log.Println("Starting gRPC server on port 50051...")
        if err := s.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    http.HandleFunc("/getProduct", func(w http.ResponseWriter, r *http.Request) {
        totalStart := time.Now()
        log.Println("Received HTTP request for /getProduct")

        productIdStr := r.URL.Query().Get("id")
        productId, err := strconv.Atoi(productIdStr)
        if err != nil {
            http.Error(w, "Invalid product ID", http.StatusBadRequest)
            return
        }

        req := &catalog.GetProductByIdRequest{Id: int32(productId)}

        grpcStart := time.Now()
        res, err := catalogClient.GetProductById(context.Background(), req)
        if err != nil {
            http.Error(w, "Failed to get product", http.StatusInternalServerError)
            return
        }
        grpcDuration := time.Since(grpcStart)

        log.Printf("Received gRPC response: %v, gRPC call took %s", res, grpcDuration)
        fmt.Fprintf(w, "Product: %v", res)

        totalDuration := time.Since(totalStart)
        httpDuration.WithLabelValues(r.URL.Path).Observe(totalDuration.Seconds())
        log.Printf("Total time taken from HTTP request to response sent: %s", totalDuration)
    })

    http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
        firstName := r.URL.Query().Get("firstName")
        lastName := r.URL.Query().Get("lastName")
        email := r.URL.Query().Get("email")
        password := r.URL.Query().Get("password")

        if firstName == "" || lastName == "" || email == "" || password == "" {
            http.Error(w, "All fields are required", http.StatusBadRequest)
            return
        }

        hashedPassword := hashPassword(password)
        authReq := map[string]string{
            "firstName": firstName,
            "lastName":  lastName,
            "email":     email,
            "password":  hashedPassword,
        }

        authReqJson, err := json.Marshal(authReq)
        if err != nil {
            http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
            return
        }

        resp, err := http.Post("http://localhost:50053/signup", "application/json", bytes.NewBuffer(authReqJson))
        if err != nil {
            http.Error(w, "Failed to communicate with auth server", http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()

        var authRes AuthResponse
        if err := json.NewDecoder(resp.Body).Decode(&authRes); err != nil {
            http.Error(w, "Failed to decode response", http.StatusInternalServerError)
            return
        }

        fmt.Fprintf(w, "Signup: %v", authRes)
    })

    http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
        email := r.URL.Query().Get("email")
        password := r.URL.Query().Get("password")

        if email == "" || password == "" {
            http.Error(w, "Email and password are required", http.StatusBadRequest)
            return
        }

        hashedPassword := hashPassword(password)
        authReq := map[string]string{"email": email, "password": hashedPassword}

        authReqJson, err := json.Marshal(authReq)
        if err != nil {
            http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
            return
        }

        resp, err := http.Post("http://localhost:50053/login", "application/json", bytes.NewBuffer(authReqJson))
        if err != nil {
            http.Error(w, "Failed to communicate with auth server", http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()

        var authRes AuthResponse
        if err := json.NewDecoder(resp.Body).Decode(&authRes); err != nil {
            http.Error(w, "Failed to decode response", http.StatusInternalServerError)
            return
        }

        fmt.Fprintf(w, "Login: %v", authRes)
    })

    http.Handle("/metrics", promhttp.Handler())

    log.Println("Starting HTTP server on port 50061...")
    if err := http.ListenAndServe(":50061", nil); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}

