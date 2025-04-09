package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/Alena-Kurushkina/gophkeeper/internal/grpc/proto"
)

type GophKeeperClient struct {
	Client pb.GophkeeperClient
	Connection *grpc.ClientConn
	Interceptor *AuthInterceptor
}

func CreateClient() *GophKeeperClient {
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

	interceptor:=&AuthInterceptor{}

    // Создание соединения
    conn, err := grpc.NewClient(
        "localhost:50051",
        grpc.WithTransportCredentials(credentials.NewTLS(creds)),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
    )
    if err != nil {
        log.Fatalf("did not connect: %v", err)
    }

	c := pb.NewGophkeeperClient(conn)

	return &GophKeeperClient{
		Client: c,
		Connection: conn,
		Interceptor: interceptor,
	}
}

type UploadResult struct{
	Status string
	FileId string
	Size int64
}

func(c *GophKeeperClient) UploadFile(ctx context.Context, filePath, metainfo string) (*UploadResult, error) {
	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Создаем стрим
	stream, err := c.Client.UploadBigData(ctx)
	if err != nil {
		return nil, err
	}

	// Отправляем метаданные первым сообщением
	err = stream.Send(&pb.UploadRequest{
		Data: &pb.UploadRequest_Metadata{
			Metadata: &pb.FileMetadata{
				Filename: file.Name(),
				Metainfo: metainfo,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Отправляем чанки (по 1 MB)
	buf := make([]byte, 1<<20) // 1 MB
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		err = stream.Send(&pb.UploadRequest{
			Data: &pb.UploadRequest_Chunk{
				Chunk: buf[:n],
			},
		})
		if err != nil {
			if err==io.EOF{
				resp, err := stream.CloseAndRecv()
				log.Print(resp)
				return nil, err
			}
			return nil, err
		}
	}

	// Получаем ответ
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		Status: resp.Status,
		Size: resp.Size,
		FileId: resp.FileId,
	}, nil
}