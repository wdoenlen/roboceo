#!/bin/bash

set -e

LOCATION="$1"
if [[ "$LOCATION" == "" ]]; then
	echo "usage: ./scrape.sh <location>" >&2
	exit 1
fi

curl \
	-d "location=$LOCATION" \
	http://localhost:9012/scrape

curl \
	-d "location=$LOCATION" \
	-d "tomorrow=true" \
	http://localhost:9012/scrape
