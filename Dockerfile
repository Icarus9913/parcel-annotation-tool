FROM golang:1.17

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY="https://goproxy.cn,direct"

WORKDIR /app/parcel-annotation-tool
COPY * ./
#RUN go mod download && go build -ldflags="-w -s" -o annotationTool main.go
RUN make all