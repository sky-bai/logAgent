FROM golang:1.16-alpine

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
COPY ./conf ./conf

RUN go build -o /logAgent



CMD [ "/logAgent" ]