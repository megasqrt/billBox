
RUN_ADDRESS=localhost:8080
DATABASE_URI=postgres://billbox:billbox@localhost:5432/billbox?sslmode=disable
ACCRUALSYSTEMADDRESS=http://localhost:8090/api/accrue
JWTSECRETKEY=jwtsecretkey 

.PHONY: echo run tests tidy build

echo:
	go version

run:
	RUN_ADDRESS=$(RUN_ADDRESS) DATABASE_URI=$(DATABASE_URI) ACCRUALSYSTEMADDRESS=$(ACCRUALSYSTEMADDRESS) JWTSECRETKEY=$(JWTSECRETKEY) go run cmd/gophermart/main.go

tests:
	go vet -vettool=$(which statictest) ./...
	go test -v -race ./...

tidy:
	go mod tidy

build: tidy tests
	go build -mod=mod -o gophermart cmd/gophermart/main.go