FROM golang:1.6.2

COPY . /go/src/github.com/maxhawkins/eventdb

RUN go get -d -v github.com/maxhawkins/eventdb/...
RUN go install -v github.com/maxhawkins/eventdb/...

CMD eventdb
