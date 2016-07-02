package scraper

import (
	"errors"
	"fmt"
	"time"

	"github.com/tebeka/selenium"
)

const scrapeScript = `
var EVENT_PATTERN = 'https:\\/\\/www\\.facebook\\.com\\/events\\/(\\d+)';
function getEventIDs() {
  var anchors = document.querySelectorAll('a');
  var idSet = {};
  for (var i = 0; i < anchors.length; i++) {
    var anchor = anchors[i];
    var re = new RegExp(EVENT_PATTERN);
    var match = anchor.href.match(re);
    if (!match) {
      continue;
    }
    var id = match[1];
    idSet[id] = true;
  }
  return Object.keys(idSet);
}
function empty(el) {
  while (el.hasChildNodes()) {
    el.removeChild(el.lastChild);
  }
}
function clearPagerContainers(el) {
  var children = el.children;
  var total = 0;
  for (var i = 0; i < children.length; i++) {
    var c = children[i];
    if (c.id.indexOf('fbBrowseScrollingPagerContainer') == 0 ||
      c.id.indexOf('browse_result_below_fold') >= 0 ||
      c.id == 'BrowseResultsContainer') {
      empty(c);
      total += 1;
      continue;
    }
    total += clearPagerContainers(c);
  }
  return total;
}
function clearAllPagerContainers() {
  var container = document.getElementById('initial_browse_result');
  clearPagerContainers(container);
}
function endOfResultsVisible() {
  var footer = document.getElementById('browse_end_of_results_footer');
  return footer !== null;
}
function scrollToBottom() {
  window.scrollTo(0, document.body.scrollHeight);
}
function scrollToTop() {
  window.scrollTo(0, 0);
}
function isLoaded() {
  if (endOfResultsVisible()) {
    return true;
  }
  var ids = getEventIDs();
  if (ids.length > 0) {
    return true;
  }
  return false;
}
function waitForLoad() {
  return new Promise(function(done) {
    var check = function() {
      if (!isLoaded()) {
        setTimeout(check, 400);
      }

      done();
    };
    check();
  });
}
function cleanup() {
  clearAllPagerContainers();
  scrollToBottom();
  scrollToTop();
}
function fetchEventPage(callback) {
  waitForLoad().then(function() {
    var ids = getEventIDs();
    cleanup();
    callback(ids);
  });
}
function fetchAllPages(callback, progress) {
	var results = [];
	function next() {
		fetchEventPage(function(result) {
			if (result.length == 0) {
				callback(results)
				return;
			}
			progress(results);
			results = results.concat(result);
			setTimeout(next, Math.random() * 500);
		});
	}
  next();
}
window.fetchEventPage = fetchEventPage;

`

type EventsPage struct {
	wd selenium.WebDriver
}

func (p *EventsPage) Verify() error {
	_, err := p.wd.FindElement(selenium.ByCSSSelector, "#BrowseResultsContainer")
	if err != nil {
		return errors.New("expected to be on events page")
	}
	return nil
}

func (p *EventsPage) InjectScraper() error {
	if _, err := p.wd.ExecuteScript(scrapeScript, []interface{}{}); err != nil {
		return err
	}

	if err := p.wd.SetAsyncScriptTimeout(3 * time.Second); err != nil {
		return err
	}

	return nil
}

func (p *EventsPage) GetEventPagelet() (ids []string, err error) {
	script := "fetchEventPage(arguments[arguments.length - 1]);"
	results, err := p.wd.ExecuteScriptAsync(script, []interface{}{})
	if err != nil {
		return nil, err
	}
	resultIDs, ok := results.([]interface{})
	if !ok {
		return nil, fmt.Errorf("bad response from js: %v", results)
	}
	for _, idInterface := range resultIDs {
		id, ok := idInterface.(string)
		if !ok {
			return nil, fmt.Errorf("bad response from js: %v", results)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func OnEventsPage(wd selenium.WebDriver) (*EventsPage, error) {
	page := &EventsPage{wd}
	if err := page.Verify(); err != nil {
		return nil, err
	}
	return page, nil
}
