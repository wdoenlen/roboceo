FROM selenium/standalone-firefox

ADD eventscrape .
ADD entrypoint.sh /entrypoint.sh
ADD www /www

EXPOSE 8080

VOLUME /db

CMD cd / && ./entrypoint.sh
