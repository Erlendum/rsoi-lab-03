# alpine build stage
FROM alpine:3.10 as alpine-build
RUN apk --no-cache add \
        ca-certificates \
        curl \
        unzip \
        make \
        wget \
        htop \
        net-tools \
        curl \
        tzdata \
        bash \
        bind-tools

# golang build stage
FROM golang:1.22 as build

ENV GO111MODULE=on

WORKDIR /go/src/app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY .. .

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/reservation-system ./cmd/reservation-system

# final stage
FROM alpine:3.10 as app

COPY --from=alpine-build .. .

WORKDIR /usr/bin

COPY --from=build /go/src/app /go
COPY ./configs configs
EXPOSE 8080

ENTRYPOINT /go/bin/reservation-system