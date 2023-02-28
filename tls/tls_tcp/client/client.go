package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

func main() {
	// client CA
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("../certs/ca.crt")
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append ca certificate")
	}

	// client certificate
	cert, err := tls.LoadX509KeyPair("../certs/client.crt", "../certs/client.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}

	// tls dial
	config := tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: false,       // 客户端对server的主机名进行校验，必须设为false避免中间人攻击
		ServerName:         "localhost", // 客户端校验server的主机名
		RootCAs:            certPool,    // 客户端校验server需用RootCAs来验证服务端的证书
		//ClientCAs:          certPool,    // 如果client也会充当server被他人请求，要校验对方也需要设置ClientCAs
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:8000", &config)
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}
	defer conn.Close()
	log.Println("client: connected to: ", conn.RemoteAddr())

	// tls state
	state := conn.ConnectionState()
	for _, v := range state.PeerCertificates {
		fmt.Println(x509.MarshalPKIXPublicKey(v.PublicKey))
		fmt.Println(v.Subject)
	}
	log.Println("client: handshake: ", state.HandshakeComplete)
	log.Println("client: mutual: ", state.NegotiatedProtocolIsMutual)

	// send recv
	message := "Hello\n"
	n, err := io.WriteString(conn, message)
	if err != nil {
		log.Fatalf("client: write: %s", err)
	}
	log.Printf("client: wrote %q (%d bytes)", message, n)

	reply := make([]byte, 256)
	n, err = conn.Read(reply)
	log.Printf("client: read %q (%d bytes)", string(reply[:n]), n)
	log.Print("client: exiting")
}
