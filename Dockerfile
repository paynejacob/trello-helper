FROM golang:1.16-alpine as build

WORKDIR /build

ADD . .

RUN go build -o trello-helper main.go
RUN chmod +x trello-helper

FROM alpine

COPY --from=build /build/trello-helper /usr/local/bin/

ENTRYPOINT ["trello-helper"]