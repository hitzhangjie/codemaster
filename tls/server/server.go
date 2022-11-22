// Go to ${grpc-up-and-running}/samples/ch02/productinfo
// Optional: Execute protoc --go_out=plugins=grpc:golang/product_info product_info.proto
// Execute go get -v github.com/grpc-up-and-running/samples/ch02/productinfo/go/product_info
// Execute go run go/server/main.go

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"time"

	"github.com/hitzhangjie/codemaster/tls/proto"
	pb "github.com/hitzhangjie/codemaster/tls/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type server struct {
	*proto.UnimplementedHelloServiceServer
}

// server is used to implement ecommerce/product_info.
func (s *server) Hello(ctx context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
	log.Printf("recv req: %+v", req)

	rsp := &proto.HelloResponse{
		ErrCode: 0,
		ErrMsg:  "success",
	}

	log.Printf("send rsp: %+v", rsp)

	return rsp, nil
}

func (s *server) mustEmbedUnimplementedHelloServiceServer() {
	panic("not implemented") // TODO: Implement
}

var (
	port    = ":50051"
	crtFile = filepath.Join("../conf/server.crt")
	keyFile = filepath.Join("../conf/server.key")
	caFile  = filepath.Join("../conf/ca.crt")
)

func main() {
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append client certs")
	}

	opts := []grpc.ServerOption{
		// Enable TLS for all incoming connections.
		grpc.Creds( // Create the TLS credentials
			credentials.NewTLS(&tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				Certificates: []tls.Certificate{certificate},
				ClientCAs:    certPool,
			})),
		grpc.ConnectionTimeout(time.Hour * 24),
	}

	s := grpc.NewServer(opts...)
	pb.RegisterHelloServiceServer(s, &server{})
	// Register reflection service on gRPC server.
	//reflection.Register(s)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening on %s", port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
