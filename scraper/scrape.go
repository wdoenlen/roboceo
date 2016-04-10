package scraper

import (
	"fmt"
	"time"

	"github.com/tebeka/selenium"
)

const MaxPages = 500

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

func login(wd selenium.WebDriver, username, password string) error {
	email, err := wd.FindElement(selenium.ById, "email")
	if err != nil {
		return fmt.Errorf("login: %v", err)
	}

	pass, err := wd.FindElement(selenium.ById, "pass")
	if err != nil {
		return fmt.Errorf("login: %v", err)
	}

	submit, err := wd.FindElement(selenium.ByCSSSelector, "#loginbutton input")
	if err != nil {
		return fmt.Errorf("login: %v", err)
	}

	if err := email.SendKeys(username); err != nil {
		return fmt.Errorf("login: %v", err)
	}

	if err := pass.SendKeys(password); err != nil {
		return fmt.Errorf("login: %v", err)
	}

	if err := submit.Click(); err != nil {
		return fmt.Errorf("login: %v", err)
	}

	return nil
}

func verifyOnWelcomePage(wd selenium.WebDriver) error {
	if _, err := wd.FindElement(selenium.ById, "loginbutton"); err == nil {
		src, _ := wd.PageSource()
		return fmt.Errorf("unexpected page source:\n%s\n", src)
	}

	return nil
}

func ensureLoggedIn(wd selenium.WebDriver, username, password string) error {
	if err := wd.Get("https://www.facebook.com"); err != nil {
		return err
	}

	err := verifyOnWelcomePage(wd)
	if err == nil { // already logged in
		return nil
	}

	if err := login(wd, username, password); err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)

	if err := verifyOnWelcomePage(wd); err != nil {
		return err
	}

	return nil
}

func GetAllEvents(wd selenium.WebDriver, searchURL, username, password string, ids chan string) error {
	if err := ensureLoggedIn(wd, username, password); err != nil {
		return err
	}

	if err := wd.Get(searchURL); err != nil {
		return err
	}

	if _, err := wd.ExecuteScript(scrapeScript, []interface{}{}); err != nil {
		return err
	}

	if err := wd.SetAsyncScriptTimeout(3 * time.Second); err != nil {
		return err
	}

	for i := 0; i < MaxPages; i++ {
		script := "fetchEventPage(arguments[arguments.length - 1]);"
		results, err := wd.ExecuteScriptAsync(script, []interface{}{})
		if err != nil {
			return err
		}
		resultIDs, ok := results.([]interface{})
		if !ok {
			return fmt.Errorf("bad response from js: %v", results)
		}
		if len(resultIDs) == 0 {
			break
		}
		for _, idInterface := range resultIDs {
			id, ok := idInterface.(string)
			if !ok {
				return fmt.Errorf("bad response from js: %v", results)
			}
			ids <- id
		}
		time.Sleep(5 * time.Second)
	}

	return nil
}
