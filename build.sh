CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a --ldflags '-extldflags "-static"' -tags netgo -installsuffix netgo -o lxcfs-initializer .
