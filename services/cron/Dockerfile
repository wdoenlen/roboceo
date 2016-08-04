FROM ubuntu:14.04

RUN sudo apt-get update && \
	sudo apt-get install rsyslog curl -y && \
	sudo apt-get clean

COPY crontab /etc/crontab
COPY run-cron.sh /usr/bin/run-cron.sh

RUN chmod +x /usr/bin/run-cron.sh
RUN touch /var/log/cron.log

CMD /usr/bin/run-cron.sh
