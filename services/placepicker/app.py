import json
import os
import random
import time

import requests

from flask import Flask, request, Response
from flask.ext.compress import Compress
from LatLon import LatLon

app = Flask(__name__)

# HACK(maxhawkins): put nginx in front of this app instead
Compress(app)
app.config['COMPRESS_MIMETYPES'] = 'application/json; charset=utf-8'

API_KEY = os.environ['PLACES_API_KEY']

with open('tabelog_data.json', 'r') as f:
	tabelog_data = json.load(f)

def random_location(lat, lng, radius_km):
	if radius_km <= 0:
		return (lat, lng)

	# TODO(maxhawkins): maybe use a gaussian instead?
	dist = random.uniform(0, radius_km)
	heading = random.uniform(0, 360)

	origin = LatLon(lat, lng)
	dest = origin.offset(heading, dist)

	return (float(dest.lat), float(dest.lon))

def nearby(lat, lng, place_types, open_now):
	params = {
		'key': API_KEY,
		'location': '%f,%f' % (lat, lng),
		# TODO(maxhawkins): this will soon be deprecated from Google Places
		'types': '|'.join(place_types),
		'rankby': 'distance',
	}
	if open_now:
		params['opennow'] = 'true'

	url = 'https://maps.googleapis.com/maps/api/place/nearbysearch/json'

	resp = requests.get(url, params=params)
	if resp.status_code != 200:
		raise StandardError('bad status %d', r.status_code)

	print(resp.url)

	data = resp.json()
	if data['status'] != 'OK':
		raise StandardError('bad status "%s"' % data['status'])

	return data['results']

def try_pick_place(src_lat, src_lng, place_types, radius_km):
	# Go a random distance in a random direction
	sample_lat, sample_lng = random_location(
		src_lat, src_lng, radius_km)

	# See what's nearby
	candidates = nearby(
		sample_lat, sample_lng, place_types, open_now=True)		

	# Remove results outside our search radius
	valid_candidates = []
	for place in candidates:
		location = place['geometry']['location']
		dest_lat, dest_lng = location['lat'], location['lng']

		origin = LatLon(src_lat, src_lng)
		dest = LatLon(dest_lat, dest_lng)

		if origin.distance(dest) > radius_km:
			continue

		valid_candidates.append(place)

	# Pick one at random
	if len(valid_candidates) == 0:
		return None

	choice = random.choice(valid_candidates)
	choice['lat'] = choice['geometry']['location']['lat']
	choice['lng'] = choice['geometry']['location']['lng']

	return choice

def pick_place(src_lat, src_lng, place_types=[], radius_km=20):
	tries = 5

	while tries > 0:
		place = try_pick_place(src_lat, src_lng, place_types, radius_km)
		if place is not None:
			return place

		tries -= 1
		app.logger.warning("no place found, trying another spot")

		# FIXME(maxhawkins): is it possible to exceed rate limit with concurrent reqs?
		time.sleep(1)

	return None

def pick_tabelog(lat, lng, radius_km=20):
	possible = []
	for place in tabelog_data:
		origin = LatLon(lat, lng)
		dest = LatLon(place['lat'], place['lng'])

		if origin.distance(dest) > radius_km:
			continue

		possible.append(place)

	if len(possible) == 0:
		return None

	return random.choice(possible)

@app.route("/destination")
def choose_destination():
	lat = float(request.args['lat'])
	lng = float(request.args['lng'])
	radius_km = float(request.args.get('radius_km', 20))

	types = []
	if 'types' in request.args:
		types = request.args['types'].split(',')
	if len(types) == 0:
		raise StandardError('at least one type required')

	dest = None
	# HACK(maxhawkins): hack in tabelog data
	if types == ['restaurant']:
		dest = pick_tabelog(lat, lng, radius_km)
	if dest is None:
		dest = pick_place(lat, lng, types, radius_km)
	if dest is None:
		return 'no results found', 500

	place = {
		'name': dest['name'],
		'lat': dest['lat'],
		'lng': dest['lng'],
	}

	js = json.dumps(place, ensure_ascii=False)

	resp = Response(js)
	resp.headers['Content-Type'] = 'application/json; charset=utf-8'
	return resp

if __name__ == "__main__":
    app.run(host='0.0.0.0', debug=True)
