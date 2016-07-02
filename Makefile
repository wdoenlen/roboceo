.PHONY: run
run: docker
	docker run -v /eventdata:/db --restart=always --name eventscrape -p 8000:8080 -d eventscrape

.PHONY: docker
docker: build/eventscrape
	docker build -t eventscrape build

build/eventscrape: $(shell find . | grep .go$)
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $@ .
