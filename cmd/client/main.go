package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	tea "github.com/charmbracelet/bubbletea"

	pb "github.com/Alena-Kurushkina/gophkeeper/internal/grpc/proto"
	"github.com/Alena-Kurushkina/gophkeeper/internal/client"
)
    // // Загрузка CA сертификата
    // caCert, err := os.ReadFile("/Users/alena/app/tls/practicum_gophkeeper_certs/ca.crt")
    // if err != nil {
    //     log.Fatal(err)
    // }

    // // Создание пула сертификатов
    // caCertPool := x509.NewCertPool()
    // caCertPool.AppendCertsFromPEM(caCert)

	// // Добавить клиентский сертификат
	// clientCert, err := tls.LoadX509KeyPair("/Users/alena/app/tls/practicum_gophkeeper_certs/client.crt", "/Users/alena/app/tls/practicum_gophkeeper_certs/client.key")
	// if err != nil {
	// 	log.Fatal(err)
	// }

    // // Настройка TLS
    // tlsConfig := &tls.Config{
	// 	RootCAs:      caCertPool,
	// 	Certificates: []tls.Certificate{clientCert},
	// }

func CreateConnection() *grpc.ClientConn {
	// Загружаем CA-сертификат
	caCert, err := os.ReadFile("/Users/alena/Library/Application Support/mkcert/rootCA.pem")
	if err != nil {
		log.Fatal("Failed to read CA cert: ", err)
	}

	// Создаём пул доверенных сертификатов
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Настраиваем TLS
	creds := &tls.Config{
		RootCAs: caCertPool,
	}

    // Создание соединения
    conn, err := grpc.NewClient(
        "localhost:50051",
        grpc.WithTransportCredentials(credentials.NewTLS(creds)),
    )
    if err != nil {
        log.Fatalf("did not connect: %v", err)
    }

	return conn
}

func CreateClient(conn *grpc.ClientConn) pb.GophkeeperClient{

    c := pb.NewGophkeeperClient(conn)

    response, err := c.CheckConnection(context.Background(), &pb.Hello{Name: "Client"})
    if err != nil {
        log.Fatalf("error calling SayHello: %v", err)
    }
    log.Printf("Response: %s", response.Response)

	return c
}

func main() {
	conn:=CreateConnection()
	defer conn.Close()
	cl:=CreateClient(conn)

	p := tea.NewProgram(client.InitialModel(cl))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Ошибка запуска программы: %v\n", err)
		os.Exit(1)
	}
	// creds:=pb.Credentials{
	// 	Login: "alena-kurushkina",
	// 	Password: "123456",
	// }

	// token, err:=cl.Register(context.Background(), &creds)
	// if err!=nil{
	// 	log.Fatalf("Error on registration: %v", err)
	// }
	// _=token
}