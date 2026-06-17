export FUNCTION_TARGET=main
export GOOGLE_APPLICATION_CREDENTIALS=key-file.json
export $(grep -v '^#' .env | xargs)
go run ./cmd/main.go
