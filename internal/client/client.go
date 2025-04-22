package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
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

func(c *GophKeeperClient) UploadFile(ctx context.Context, filePath, fileName, metainfo string) (*UploadResult, error) {
	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err:=file.Stat()
	if err!=nil{
		return nil, err
	}


	// Создаем стрим
	stream, err := c.Client.UploadBigData(ctx)
	if err != nil {
		return nil, err
	}

	// Отправляем метаданные первым сообщением
	err = stream.Send(&pb.BigDataChunk{
		Data: &pb.BigDataChunk_Metadata{
			Metadata: &pb.FileMetadata{
				Filename: fileName,
				Metainfo: metainfo,
				Size:  stat.Size(),
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

		err = stream.Send(&pb.BigDataChunk{
			Data: &pb.BigDataChunk_Chunk{
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

type DownloadResult struct {
	Filename string
	Size int64
	Metainfo string
}

func(c *GophKeeperClient) DownloadFile(ctx context.Context, filename, outputPath string) (*DownloadResult, error) {
	stream, err := c.Client.DownloadBigData(ctx, &pb.DownloadRequest{
        Filename: filename,
    })
    if err != nil {
        return nil, err
    }

    // Получаем метаданные
    firstChunk, err := stream.Recv()
    if err != nil {
        return nil, fmt.Errorf("failed to receive metadata: %v", err)
    }

    metadata := firstChunk.GetMetadata()
    if metadata == nil {
        return nil, errors.New("first message must contain metadata")
    }

    file, err := os.Create(outputPath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    log.Printf(
        "Downloading file: %s, size %d",
        metadata.Filename,
		metadata.Size,
    )

    // Получаем и записываем чанки
    var receivedBytes int64
    for {
        chunk, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }

        content := chunk.GetChunk()
        if content == nil {
            continue // Пропускаем сообщения без контента
        }

        n, err := file.Write(content)
        if err != nil {
            return nil, err
        }
        receivedBytes += int64(n)
    }

    // Проверяем целостность файла
    if receivedBytes != metadata.Size {
        return nil, fmt.Errorf(
            "file size mismatch: expected %d, received %d",
            metadata.Size,
            receivedBytes,
        )
    }

    log.Println("File downloaded successfully")
    return &DownloadResult{
		Filename: metadata.Filename,
		Size: receivedBytes,
		Metainfo: metadata.Metainfo,
	}, nil
}