package grpctls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	maxRetry      = 20
	sleepDuration = 10 * time.Millisecond
)

type Client interface {
	IsAlive() bool          // Check if the connection is alive
	Close() (string, error) // Close safely the connection
}
type GrpcClient struct {
	Host              string
	Port              string
	Conn              *grpc.ClientConn
	cachedCertificate *x509.Certificate // Move the declaration here
	config            *tls.Config
}

// NewClientWithContextTLS creates a new gRPC client with TLS support.
func NewClientWithContextTLS(ctx context.Context, host, port string, dialOptions []grpc.DialOption) (*GrpcClient, error) {
	if host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}
	if port == "" {
		return nil, fmt.Errorf("port cannot be empty")
	}

	address := fmt.Sprintf("%s:%s", host, port)

	var err error
	var cachedCertificate *x509.Certificate // Define a local variable

	for i := 0; i < maxRetry; i++ {
		cachedCertificate, err = fetchCertificate(address)
		if err == nil {
			break
		}

		log.Printf("error fetching certificate: %v\n", err)
		time.Sleep(sleepDuration)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(cachedCertificate)
	config := &tls.Config{RootCAs: caCertPool}

	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(config))}
	opts = append(opts, dialOptions...)

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %v", err)
	}

	return &GrpcClient{
		Host:              host,
		Port:              port,
		Conn:              conn,
		cachedCertificate: cachedCertificate, // Assign the local variable to the struct field
		config:            config,
	}, nil
}

// Close safely closes the connection.
func (c *GrpcClient) Close() (string, error) {
	if c.Conn == nil {
		return "", fmt.Errorf("connection is nil")
	}
	return "Connection closed", c.Conn.Close()
}
