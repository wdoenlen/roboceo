FROM selenium/standalone-firefox

ADD eventscrape .
ADD entrypoint.sh /entrypoint.sh

EXPOSE 8080

VOLUME /db

CMD cd / && ./entrypoint.sh
