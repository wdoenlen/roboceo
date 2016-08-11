FROM ubuntu:14.04

RUN sudo apt-get update && \
	sudo apt-get install rsyslog curl -y && \
	sudo apt-get clean

COPY crontab /etc/crontab
RUN touch /var/log/cron.log

COPY run-cron.sh /usr/bin/run-cron.sh
RUN chmod +x /usr/bin/run-cron.sh

COPY scrape-events.sh /usr/bin/scrape-events.sh
RUN chmod +x /usr/bin/scrape-events.sh

CMD /usr/bin/run-cron.sh
