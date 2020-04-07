install:bindata
	go install
build: bindata
	go build -o ttool ./
bindata:
	cd template; go-bindata -o ../pkg/bindata.go -pkg pkg  ./...;
