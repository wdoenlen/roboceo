"""
Documentation on googleplaces: https://github.com/slimkrazy/python-google-places

TODO:
    - Integrate with scheduler so that scheduler picks place types and then
      feeds them into PlacePicker
    - Pick radius from a given probability distribution, but this should be
      fed into PlacePicker as well

"""

import random
from googleplaces import GooglePlaces, types, ranking

API_KEY = ""

class PlacePicker(object):

    def __init__(self, lat, lng, radius=1500, place_types=[], places_api=GooglePlaces(API_KEY)):
        self.lat = lat
        self.lng = lng
        self.radius = radius
        self.place_types = place_types
        self.google_places = places_api
        self.next_destination = None

    def get_nearby(self, rankby=None):
        if self.place_types and rankby in [ranking.DISTANCE, ranking.PROMINENCE]:
            return self.google_places.nearby_search(lat_lng={"lat": self.lat,
                                                            "lng": self.lng},
                                                    radius=self.radius,
                                                    types=self.place_types,
                                                    rankby=rankby)
        elif self.place_types:
            return self.google_places.nearby_search(lat_lng={"lat": self.lat,
                                                            "lng": self.lng},
                                                    radius=self.radius,
                                                    types=self.place_types)
        else:
            return self.google_places.nearby_search(lat_lng={"lat": self.lat,
                                                            "lng": self.lng},
                                                    radius=self.radius)

    def pick_next_destination(self):
        nearby_places = self.get_nearby()
        if len(nearby_places.places) > 0:
            self.next_destination = random.choice(nearby_places.places)
        return self.next_destination
