FROM golang
ADD . /go/src/github.com/maxhawkins/eventscrape
RUN go get github.com/maxhawkins/eventscrape
