package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

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

func main() {

	cl:=client.CreateClient()
	defer cl.Connection.Close()

	//menu.RunMainMenu()

	p := tea.NewProgram(client.InitialModel(cl))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Ошибка запуска программы: %v\n", err)
	}
}