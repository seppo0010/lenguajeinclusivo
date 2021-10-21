FROM golang:1.17

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["crawler"]
