FROM golang:1.17.2-bullseye as builder
WORKDIR /go/src/app

COPY ./builder/go.mod ./builder/go.sum ./
RUN go mod download

COPY ./builder .

RUN go get -d -v ./...
RUN go build .


FROM node:14-buster-slim as build
WORKDIR /app

RUN apt-get update && apt-get install -y \
    poppler-utils \
    tesseract-ocr-spa \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

ENV PATH /app/node_modules/.bin:$PATH
COPY web/package.json ./
#COPY web/yarn.lock ./
COPY web/ts/package.json ./ts/
#COPY web/ts/yarn.lock ./ts/
RUN yarn

WORKDIR /app/ts
RUN yarn

WORKDIR /app
COPY web/ ./

COPY --from=builder /go/src/app/juscaba /app/builder
COPY run.sh /app/
