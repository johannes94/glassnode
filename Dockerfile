FROM golang:1.17.3 as build_stage

RUN mkdir /src
WORKDIR /src
COPY . .
RUN go build -o /src/api

ENTRYPOINT ["/src/api"]
