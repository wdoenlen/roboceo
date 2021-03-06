import datetime
import json
import time
import random

SCRAPE_JS = '''
// This file contains the JavaScript source for a Google Maps distance
// data scraper. As of July 23, 2016 it works with the running version
// of Google Maps.
// 
// It exploits the distance as you drag feature on the directions page
// for Google Maps. It navigates to a URL that displays directions starting
// from a selected origin and then simulates mouse click and drag events
// at random points on the map canvas.
// 
// Right now it's just a proof of concept. Soon I will wrap it in a webdriver
// so that it can be used headless by a batch job. This will allow us to quickly
// create [isochrones](https://en.wikipedia.org/wiki/Isochrone_map) from any
// point on the globe and therefore efficiently estimate which candidate points
// are accessible within a given time window in apps like placepicker and tsp.
// 
// To run, execute the BuildURL function below and navigate to the generated URL.
// Then copy-paste the rest of the source into the console and call StartScrape().

window.TODO = [];
window.RESULTS = [];

// when set to false in the console, this stops
// the loop function and halts the scrape.
var stop = false;

// StartScrape kicks things off! Only call this after you're on the page
// returned by BuildURL.
window.StartScrape = function() {
  var page = new DirectionsPage();
  page.reverseDirections()
    .then(function() {
      return page.doInitialClick();
    })
    .then(function() {

      function loop() {
        // For now, just choose random offsets in screen space. Later
        // we might be interested in specific lat-lngs.
        var point = window.TODO.pop();
        if (!point) {
          sleep(500).then(loop);
        }

        page.sampleMap(point)
          .then(function(result) {
            console.log('result = ', result); // it worked!

            RESULTS.push({
              point: point,
              info: result,
            });

            if (!stop) {
              loop();
            }
          }, function() {
            // TODO(maxhawkins): track down this bug
            // 
            // For some reason the trip info box doesn't change after you move
            // the mouse. Right now I'm just ignoring it, waiting a while and moving
            // on. Later I'd like to get to the bottom of it. I suspect it has
            // something to do with mousemove-ing to points that are nearby the previous
            // point we moved to.
            console.warn('trip info never appeared. this is a bug. point = ', point);

            RESULTS.push({
              point: point,
              error: 'trip info never appeared',
            });

            if (!stop) {
              loop();
            }
          });
      }
      loop();

    });
}


// This class contains methods for scraping the travel time information
// on http://www.google.com/maps/dir.
function DirectionsPage() {
  this.root = document;
  this.coords = new Coords(this.root);

  this.checkInInitState();

  this.state = 'init';
}

// Makes sure that we're on the directions page and the source hasn't
// been filled out yet. The page needs to be in this state for
// doInitialClick to work.
DirectionsPage.prototype.checkInInitState = function() {
  var originBox = this.root.querySelector('#directions-searchbox-0 .tactile-searchbox-input');
  var destBox = this.root.querySelector('#directions-searchbox-1 .tactile-searchbox-input');

  if (!originBox || originBox.value !== '') {
    throw new Error('expected origin not to be set in init state');
  }
  if (!destBox || destBox.value === '') {
    throw new Error('expected destination to be set in init state');
  }
};

// Gets the contents of the 'dragging-trip-info' box which gives
// travel time estimates while dragging the mouse.
DirectionsPage.prototype.getTripInfo = function() {
  var titleDiv = this.root.querySelector('.dragging-trip-title');
  var subtitleDiv = this.root.querySelector('.dragging-trip-subtitle');
  var modeDiv = this.root.querySelector('.dragging-travel-mode');
  if (!titleDiv || !subtitleDiv || !modeDiv) {
    return null;
  }
  return {
    title: titleDiv.textContent,
    subtitle: subtitleDiv.textContent,
    mode: modeDiv.textContent,
  };
};

// Returns a promise that resolves with the trip info when it has
// appeared on the page. Used to wait for the box when the page
// first loads.
DirectionsPage.prototype.untilTripInfoAppears = function() {
  var that = this;

  return new Promise(function(resolve, reject) {
    function check() {
      var info = that.getTripInfo();
      if (info) {
        resolve(info);
        clearInterval(interval);
      }
    }
    ;
    var interval = setInterval(check, 200);
    check();
  });
}

// Returns a promise that resolves with the tripInfo the next
// time it is updated in the DOM. Used to watch for updates after
// moving the mouse.
DirectionsPage.prototype.untilTripInfoChange = function() {
  var infoDiv = document.querySelector('.dragging-trip-info');

  if (!infoDiv) {
    return this.untilTripInfoAppears();
  }

  var that = this;
  return new Promise(function(resolve, reject) {

    function listener() {
      infoDiv.removeEventListener('DOMSubtreeModified', listener);

      // Wait for things to settle down before resolving
      setTimeout(function() {
        var info = that.getTripInfo();
        resolve(info);
      }, 200);
    }
    infoDiv.addEventListener('DOMSubtreeModified', listener);

    // Something is wrong if it takes too long
    setTimeout(reject, 2000);

  });
};

// sendMouseEvent sends a fake mouse event of a given type to the
// canvas element. It takes a Point object specifying where to click
// (see Coord).
DirectionsPage.prototype.sendMouseEvent = function(type, point, target) {
  var pixel = this.coords.toWindowPixel(point);
  var x = pixel.x;
  var y = pixel.y;

  var canvas = target || this.root.querySelector('canvas.widget-scene-canvas');
  var evt = new MouseEvent(type, {
    view: window,
    bubbles: true,
    cancelable: true,
    offsetX: x,
    offsetY: y,
    clientX: x,
    clientY: y,
    pageX: x,
    pageY: y,
    screenX: x,
    screenY: y,
    shiftKey: false,
  });
  canvas.dispatchEvent(evt);
};

// sampleMap moves to the given Point and resolves once the reported
// distance changes.
// 
// This function must be called in 'dragging' mode meaning it's not
// concurrent safe and you can't call it while you're waiting for
// results from a previous call to sampleMap.
DirectionsPage.prototype.sampleMap = function(point) {
  if (this.state !== 'dragging') {
    throw new Error('sampleMap from invalid state ' + this.state);
  }
  this.state = 'loading';

  var promise = this.untilTripInfoChange();

  this.sendMouseEvent('mousemove', point);

  var resetState = (function() {
    this.state = 'dragging';
  }).bind(this);
  promise.then(resetState, resetState);

  return promise;
};

// sleep is a Promise version of the sleep() function. It
// resolves after a timeout (in milliseconds).
function sleep(timeout) {
  return new Promise(function(resolve) {
    setTimeout(resolve, timeout);
  });
}

DirectionsPage.prototype.reverseDirections = function() {
  var button = document.querySelector('.widget-directions-reverse');

  this.sendMouseEvent('click', {
    top: 0,
    left: 0
  }, button);

  return sleep(500);
}

// doInitialClick puts the directions page into 'distance on drag'
// mode. It does this by clicking at an arbitrary point (to enable
// directions), waiting a while for the mode to enable, and pressing
// down again to initiate a drag.
// 
// It assumes you're in 'init' mode meaning you're at the directions
// page but haven't selected an origin yet.
DirectionsPage.prototype.doInitialClick = function(callback) {
  if (this.state !== 'init') {
    throw new Error('doInitialClick called after init');
  }

  // you gotta start somewhere
  var topLeft = {
    top: 0,
    left: 0
  };

  this.sendMouseEvent('mousedown', topLeft);

  return sleep(500)
    .then(function() {
      this.sendMouseEvent('mouseup', topLeft);

      // TODO(maxhawkins): verify it actually worked instead of just waiting
      return sleep(5000);

    }.bind(this))
    .then(function() {
      this.sendMouseEvent('mousedown', topLeft);

      return sleep(500);

    }.bind(this))
    .then(function() {
      this.state = 'dragging';

    }.bind(this));
};

// This scraper uses three coordinate systems. I'd like to make
// it easy to convert between them. This class allows you to accept
// arguments in whatever system you like and convert them to the
// proper system before use.
// 
// There are three coordinate systems and therefore three ways
// to specify a point
// 
// == Latitude/Longitude ==
// This is used for specifying geographic coordinates on the map.
// It links our click events to real places on the map. Points
// in this system look like {lat: 80, lng: 100}.
// 
// == Rect ==
// It's important to know the clickable bounds of the map object
// because if we click outside of them it messes up the scrape.
// Therefore I made a coordinate system called 'rect' where 0,0 is
// the top/left of the clickable area and 1,1 is the bottom right.
// These sorts of coordinates look like {top: 0.1, left: 0.2}.
// 
// == Window Pixel ==
// This is the offset in pixels from the top/left of the browser
// window. We use this for calculating click coordinates. Point
// objects in this system look like {x: 500, y: 600}. The range
// of x is 0..window width and the range of y is 0..window height.
function Coords() {
  var omnibox = document.querySelector('#omnibox');
  this.sidebarWidth = omnibox.getBoundingClientRect().width;
  this.scrollPadding = 50;
}

Coords.prototype.toWindowPixel = function(point) {
  if (point.top !== undefined) {
    var canvasX = this.sidebarWidth + this.scrollPadding;
    var canvasY = this.scrollPadding;
    var canvasHeight = window.innerHeight - this.scrollPadding * 2;
    var canvasWidth = window.innerWidth - this.scrollPadding * 2 - this.sidebarWidth;
    var x = canvasX + point.left * canvasWidth;
    var y = canvasY + point.top * canvasHeight;

    return {
      x: x,
      y: y
    };
  } else if (point.lat !== undefined) {
    throw new Error('unimplemented');
  } else if (point.x !== undefined) {
    return point;
  } else {
    throw new Error('unknown coordinate type');
  }
};

Coords.prototype.toLatLng = function(point) {
  if (point.top !== undefined) {
    throw new Error('unimplemented');
  } else if (point.lat !== undefined) {
    return point;
  } else if (point.x !== undefined) {
    throw new Error('unimplemented');
  } else {
    throw new Error('unknown coordinate type');
  }
};

Coords.prototype.toRect = function(point) {
  if (point.top !== undefined) {
    return point;
  } else if (point.lat !== undefined) {
    throw new Error('unimplemented');
  } else if (point.x !== undefined) {
    throw new Error('unimplemented');
  } else {
    throw new Error('unknown coordinate type');
  }
};
'''

class Scraper(object):
  def __init__(self, browser, lat, lng):
    self.browser = browser
    self._setup(lat, lng)

  def get_remaining(self):
    '''
    Returns the number of points in the queue that remain to
    be sampled. When this reaches zero we're done.
    '''
    return self.browser.execute_async_script('''
      var callback = arguments[arguments.length - 1];
      var remaining = window.TODO.length;
      callback(remaining);
    ''')

  def get_results(self):
    '''
    Remove all of the scrape results waiting in the browser's
    completed queue and return them. Returns a generator.
    '''
    results = self.browser.execute_async_script('''
      var callback = arguments[arguments.length - 1];
      var results = window.RESULTS;
      window.RESULTS = [];
      callback(results);
    ''')
    for res in results:
      try:
        left = res['point']['left']
        top = res['point']['top']
        duration = int(res['info']['title'].replace(' min', ''))
        yield "%f %f %f" % (left, top, duration)
      except:
        pass

  def _setup(self, lat, lng):
      # TODO(maxhawkins): at some point we might want to sample
      # the map more intelligently (more density at transition
      # points for instance). For now, just build a 32x32 grid.
      todo = []
      for x in range(0, 32):
        for y in range(0, 32):
          todo.append({
            'left': float(x) / 32,
            'top': float(y) / 32
          })

      # HACK(maxhawkins): shuffling the list avoids a problem
      # where the map doesn't update if you sample two places
      # in a row that are very close to each other. This should
      # be fixed in the JS scraper code and this shuffle should
      # be removed.
      random.shuffle(todo)

      url = build_url(lat, lng)
      self.browser.get(url)

      self.browser.execute_script(SCRAPE_JS)
      self.browser.execute_script('StartScrape()')

      self.browser.execute_script(
        '%s.forEach(function(t) {window.TODO.push(t) })' % json.dumps(todo))

# build_url outputs a Google maps URL for a map centered on the given point.
# After navigating to this page you can use StartScrape() to extract distances
# from the given point to other places near it.
def build_url(lat_src, lng_src):
	lat_center = lat_src
	lng_center = lng_src
	zoom = 9

	epoch = datetime.datetime.utcfromtimestamp(0)
	start_utc = int((datetime.datetime.now() - epoch).total_seconds())

	return ''.join([
		"https://www.google.com",
		"/maps",
		"/dir",
		"//'%f,%f'" % (lat_src, lng_src),
		"/@%f,%f,%dz" % (lat_center, lng_center, zoom),
		"/data=!3m1!4b1!4m11!4m10!1m0!1m3!2m2!1d%f!2d%f!2m3!6e0!7e2!8j%d!3e3" % (lat_src, lng_src, start_utc),
	])
