package main

import (
    "context"
    "log"
    "net"
    "net/http"
    "time"

    "github.com/goperfapps/microservices/catalog"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "google.golang.org/grpc"
)

var (
    grpcDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "grpc_request_duration_seconds",
        Help:    "Duration of gRPC requests.",
        Buckets: prometheus.DefBuckets,
    }, []string{"method"})
)

func init() {
    prometheus.MustRegister(grpcDuration)
}

type server struct {
    catalog.UnimplementedCatalogServiceServer
}

func (s *server) GetProductById(ctx context.Context, req *catalog.GetProductByIdRequest) (*catalog.Product, error) {
    start := time.Now()
    log.Printf("Received request for Product ID: %d", req.Id)

    // Simulate fetching product from a database or external service
    time.Sleep(2 * time.Second)

    product := &catalog.Product{
        Id:    req.Id,
        Name:  "Sample Product",
        Price: 9.99,
    }

    duration := time.Since(start)
    grpcDuration.WithLabelValues("GetProductById").Observe(duration.Seconds())
    log.Printf("Responding with product: %v, took %s", product, duration)
    return product, nil
}

func (s *server) ListProducts(ctx context.Context, req *catalog.ListProductsRequest) (*catalog.ListProductsResponse, error) {
    // Simulate fetching a list of products (could be from a database or external service)
    products := []*catalog.Product{
        {Id: 1, Name: "Product 1", Price: 19.99},
        {Id: 2, Name: "Product 2", Price: 29.99},
        {Id: 3, Name: "Product 3", Price: 39.99},
    }

    return &catalog.ListProductsResponse{
        Products: products,
    }, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50052")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    s := grpc.NewServer()
    catalog.RegisterCatalogServiceServer(s, &server{})
    log.Println("Starting gRPC server on port 50052...")
    go func() {
        if err := s.Serve(lis); err != nil {
            log.Fatalf("Failed to serve gRPC server: %v", err)
        }
    }()

    // HTTP server for Prometheus metrics
    http.Handle("/metrics", promhttp.Handler())
    log.Println("Starting metrics HTTP server on port 9091...")
    go func() {
        if err := http.ListenAndServe(":9091", nil); err != nil {
            log.Fatalf("Failed to serve metrics: %v", err)
        }
    }()

    // Graceful shutdown handling
    // ...

    // Keep the main goroutine running until interrupted
    select {}
}

