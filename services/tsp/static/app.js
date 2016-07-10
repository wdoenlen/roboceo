'use strict';

// Edit these variables to change the optimization:
var start = Date.now() / 1000; //new Date() / 1000;
var PARAMS = {
  origin: {
    lat: 40.8274149,
    lng: 140.6929728,
  },
  start: start,
  end: start + 60 * 60 * 24,
};

function Map() {
  this.el = document.createElement('div');
  this.el.className = 'map';
  this.map = new google.maps.Map(this.el, {
    zoom: 14,
    center: PARAMS.origin,
  });
}

Map.prototype.render = function(data) {
  var nodes = data.map(function(result) {
    return result.node;
  });

  var bounds = new google.maps.LatLngBounds();
  var path = [];
  nodes.forEach(function(node, i) {
    var position = new google.maps.LatLng(
      node.lat,
      node.lng
    );
    bounds.extend(position);
    path.push(position);

    var color = 'red';
    if (i == 0 || i == nodes.length - 1) {
      color = 'black';
    }

    var marker = new google.maps.Marker({
      position: position,
      map: this.map,
      icon: {
        path: google.maps.SymbolPath.CIRCLE,
        scale: 6,
        fillColor: color,
        fillOpacity: 1,
        strokeWeight: 0,
      },
    });
    marker.addListener('click', function() {
      link.href = 'http://maps.google.com/?q=' + node.lat + ',' + node.lng;
      open(url);
    });
  }.bind(this));
  this.map.fitBounds(bounds);

  var polyline = new google.maps.Polyline({
    path: path,
    strokeColor: '#0000FF',
    strokeOpacity: 1.0,
    strokeWeight: 2,
    map: this.map
  });
};

function shuffle(a) {
  var j;
  var x;
  var i;
  for (i = a.length; i; i -= 1) {
    j = Math.floor(Math.random() * i);
    x = a[i - 1];
    a[i - 1] = a[j];
    a[j] = x;
  }
}

function pad(str) {
  return ('00000' + str).slice(-2);
}

function timestamp(date) {
  return pad(date.getHours()) + ':' + pad(date.getMinutes());
}

// Generates a div with start/end times and links pointing to
// each event in the result list. Really basic for now. Later
// I'd like to link them to points on the map on hover.
function renderResults(results) {
  var div = document.createElement('div');

  for (var i = 0; i < results.length; i++) {
    var res = results[i];
    var start = new Date(res.start * 1000);
    var end = new Date(res.end * 1000);
    var name = '';
    var event = res.node.event;
    var name = res.node.name;
    if (i == 0 || i == results.length - 1) {
      name = 'Home';
    }
    var row = document.createElement('div');
    var time = document.createElement('span');
    time.innerText = '' + timestamp(start) + '-' + timestamp(end) + '\t';
    row.appendChild(time);
    var link = document.createElement('a');
    link.target = '_blank';
    link.href = 'http://maps.google.com/?q=' + res.node.lat + ',' + res.node.lng;
    link.innerText = name;
    row.appendChild(link);

    div.appendChild(row);
  }
  ;

  return div;
}

function isGoodEvent(event) {
  // Location and time is required for routing.
  // Discard events that don't have both.
  if (!event.start_time ||
    !event.end_time ||
    !event.place ||
    !event.place.location ||
    !event.place.location.latitude) {
    return false;
  }

  var idsToSkip = [
    '1737595406527624', // example of skipping an event
    '188888561499126',
  ];
  for (var i = 0; i < idsToSkip.length; i++) {
    if (event.id === idsToSkip[i]) {
      return false;
    }
  }

  var ownersToSkip = [
    'Bandsintown', // Low-quality machine generated events
  ];
  if (event.owner) {
    var owner = event.owner.name;
    for (var i = 0; i < ownersToSkip.length; i++) {
      if (owner === ownersToSkip[i]) {
        return false;
      }
    }
  }

  var start = Date.parse(event.start_time);
  var end = Date.parse(event.end_time);
  // Events that last longer than a day are usually spammy
  if ((end - start) > 1000 * 60 * 60 * 24) {
    return false;
  }

  return true;
}

// Filter events in the same place with nearly the
// same start time. These are likely duplicate listings.
// Assumes that each event has a start/end time and a
// street address.
function filterNearDuplicates(events) {
  var filtered = [];

  for (var i = 0; i < events.length; i++) {
    var event = events[i];

    var start = Date.parse(event.start_time);
    var end = Date.parse(event.end_time);
    var street = event.place.location.street;

    var hasDuplicate = false;
    for (var j = 0; j < filtered.length; j++) {
      var existing = filtered[j];
      var existingStreet = existing.place.location.street;
      var existingStart = Date.parse(existing.start_time);

      var sameStreet = existingStreet === street;
      var similarStart = Math.abs(existingStart - start) < 1000 * 60 * 60;

      if (sameStreet && similarStart) {
        hasDuplicate = true;
        break;
      }
    }

    if (hasDuplicate) {
      continue;
    }

    filtered.push(event);
  }

  return filtered;
}

function fetchAomoriNodes() {
  return fetch('aomori.json')
    .then(function(resp) {
      if (resp.status < 200 || resp.status >= 300) {
        var msg = 'bad response (' + resp.status + ')';
        return Promise.reject(msg)
      }

      return resp.json();
    })
    .then(function(nodes) {
      shuffle(nodes);

      var selected = [];
      for (var i = 0; i < nodes.length; i++) {
        var node = nodes[i];
        if (node.lng > 140.689545 && node.lng < 140.83786 && node.lat > 40.771702 && node.lat < 40.837191) {
          continue;
        }

        if (Math.random() > 0.5) {
          selected.push(nodes[i]);
        }
      }

      return selected;
    });
}

// Get the events dump from fb_data.json, filter any unusable
// data, and return a set of node objects to end to the
// routing algorithm.
function fetchEventNodes(url) {
  return fetch(url)
    .then(function(resp) {
      if (resp.status < 200 || resp.status >= 300) {
        var msg = 'bad response (' + resp.status + ')';
        return Promise.reject(msg)
      }

      return resp.json();
    })
    .then(function(body) {

      var filtered = body.filter(isGoodEvent);
      filtered = filterNearDuplicates(filtered);

      var nodes = filtered.map(function(event) {
        var loc = event.place.location;

        var start = Date.parse(event.start_time) / 1000;
        var end = Date.parse(event.end_time) / 1000;

        var node = {
          'lat': loc.latitude,
          'lng': loc.longitude,
          'start': start,
          'end': end,
          'event': event,
        };

        return node;
      })

      shuffle(nodes);
      return nodes;
    });
}


// Send a set of nodes to the Python backend and
// return the computed TSP route.
function computeRoute(nodes) {
  return fetch('/routes', {
    'method': 'POST',
    'headers': {
      'Accept': 'application/json',
      'Content-Type': 'application/json'
    },
    'body': JSON.stringify({
      'nodes': nodes,
      'startTime': PARAMS.start,
      'endTime': PARAMS.end,
      'start': PARAMS.origin,
    }),
  })
    .then(function(resp) {
      if (resp.status < 200 || resp.status >= 300) {
        var msg = 'bad response (' + resp.status + ')';
        return Promise.reject(msg)
      }

      return resp.json();
    });
}

function initMap() {
  var map = new Map();
  document.getElementById('main').appendChild(map.el);

  var $sidebar = document.getElementById('sidebar');
  $sidebar.innerText = 'Loading. This\'ll take a while...';

  // fetchEventNodes('fb_data.json')
  fetchAomoriNodes()
    .then(computeRoute)
    .then(function(body) {
      var results = body.results;

      $sidebar.innerHTML = '';
      var row = renderResults(results);
      $sidebar.appendChild(row);

      map.render(results);
    });
}
;
