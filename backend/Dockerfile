FROM golang:latest

WORKDIR /go/src/app
COPY . .

RUN go get github.com/pilu/fresh

CMD [ "fresh" ]
