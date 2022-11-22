// Go to ${grpc-up-and-running}/samples/ch02/productinfo
// Optional: Execute protoc --go_out=plugins=grpc:golang/product_info product_info.proto
// Execute go get -v github.com/grpc-up-and-running/samples/ch02/productinfo/golang/product_info
// Execute go run go/client/main.go

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"google.golang.org/grpc/credentials"

	pb "github.com/hitzhangjie/codemaster/tls/proto"
	"google.golang.org/grpc"
)

var (
	address  = "localhost:50051"
	hostname = "localhost"
	crtFile  = filepath.Join("../conf/client.crt")
	keyFile  = filepath.Join("../conf/client.key")
	caFile   = filepath.Join("../conf/ca.crt")
)

func main() {
	// Load the client certificates from disk
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("could not load client key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	// Append the certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append ca certs")
	}

	opts := []grpc.DialOption{
		// transport credentials.
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			ServerName:   hostname, // NOTE: this is required!
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: time.Hour * 24,
		}),
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewHelloServiceClient(conn)

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	ctx := context.Background()

	req := &pb.HelloRequest{Msg: "hello world"}

	rsp, err := c.Hello(ctx, req)
	if err != nil {
		log.Fatalf("say hello to server err: %+v", err)
	}
	log.Printf("say hello to server ok, req: %+v, rsp: %+v", req, rsp)
}
