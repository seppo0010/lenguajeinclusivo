FROM node:slim as build
WORKDIR /app
ENV PATH /app/node_modules/.bin:$PATH
COPY web/package.json ./
COPY web/yarn.lock ./
RUN yarn
COPY web/ ./
RUN yarn run build

FROM golang:1.17.2-bullseye
WORKDIR /go/src/app

RUN apt-get update && apt-get install -y \
    poppler-utils \
    tesseract-ocr-spa \
    && rm -rf /var/lib/apt/lists/*


COPY ./builder/go.mod ./builder/go.sum ./
RUN go mod download

COPY ./builder .

RUN go get -d -v ./...
RUN go install -v ./...

COPY --from=build /app/build /app/web

CMD ["juscaba"]
