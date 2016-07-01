from placepicker import PlacePicker

from googleplaces import GooglePlaces
import json
import os

from flask import Flask, request
app = Flask(__name__)

places_api = GooglePlaces(os.environ['PLACES_API_KEY'])

@app.route("/")
def hello():
	lat = float(request.args['lat'])
	lng = float(request.args['lng'])

	types = []
	if 'types' in request.args:
		types = request.args['types'].split(',')

	picker = PlacePicker(
		lat, lng, place_types=types, places_api=places_api)

	dest = picker.pick_next_destination()

	return json.dumps({
		'name': dest.name,
		'lat': float(dest.geo_location['lat']),
		'lng': float(dest.geo_location['lng']),
	})

if __name__ == "__main__":
    app.run(host='0.0.0.0')
