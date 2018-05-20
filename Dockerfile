FROM golang:alpine as build
MAINTAINER Eugene tsarikau@gmail.com

RUN apk add --update curl git
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/app

COPY . .

RUN dep ensure
RUN go test
RUN go build -o build/app 


FROM alpine
WORKDIR /root
COPY --from=build /go/src/app/build/app .
EXPOSE 8080
CMD ./app