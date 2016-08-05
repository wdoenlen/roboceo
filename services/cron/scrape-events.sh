#!/bin/bash

LOCATION="$1"

LOGFILE="/var/log/cron.log"

ERR=`curl -s -S -d "location=$LOCATION" -X POST http://eventscrape:8080 2>&1`

if [[ $? -eq 0 ]]; then
	echo "Scraping '$LOCATION'..." >> "$LOGFILE"
else
	echo "[error] can't scrape '$LOCATION': $ERR" >> "$LOGFILE"
fi
