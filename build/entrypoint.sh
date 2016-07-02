#!/bin/bash

# TODO(maxhawkins): what if it dies?
/opt/bin/entry_point.sh &

sudo /eventscrape -db /db/app.db
