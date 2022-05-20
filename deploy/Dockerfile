FROM golang:1.18-alpine

WORKDIR /opt/tinkoff-invest-robot-contest-2022
COPY . .
RUN go install ./cmd/trading-robot
ENTRYPOINT /go/bin/trading-robot

EXPOSE 2112
