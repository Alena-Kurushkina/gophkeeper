# openssl genrsa -out ca.key 4096
# openssl req -new -x509 -days 365 -key ca.key -out ca.crt

# # server
# openssl genrsa -out server.key 4096
# openssl req -new -key server.key -out server.csr -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost,IP:127.0.0.1,IP:::1"
# openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out server.crt -extfile <(printf "subjectAltName=DNS:localhost,IP:127.0.0.1,IP:::1")

# # client
# openssl genrsa -out client.key 4096
# openssl req -new -key client.key -out client.csr -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost,IP:127.0.0.1,IP:::1"
# openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -set_serial 02 -out client.crt -extfile <(printf "subjectAltName=DNS:localhost,IP:127.0.0.1,IP:::1")


protoc --go_out=./internal/grpc --go_opt=paths=source_relative --go-grpc_out=./internal/grpc --go-grpc_opt=paths=source_relative proto/gophkeeper.proto

# cd /Users/alena/app/tls/practicum_gophkeeper_certs

brew install mkcert  # macOS
mkcert -install
mkcert localhost 127.0.0.1 ::1

# add env var for cert path and cert key path

mockgen -destination=internal/grpc/api/mock_store.go -source=internal/gophkeeper/core.go Storager