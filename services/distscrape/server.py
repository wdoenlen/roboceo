import scraper

import time

from tornado import gen 
from tornado.ioloop import IOLoop
import tornado.options
import tornado.web

class ScrapeHandler(tornado.web.RequestHandler):
	'''
	ScrapeHandler is an HTTP handler that runs the Google Maps distance
	scraper in scraper.py and returns the results as the response.

	Use it like this:
	curl -d 'location=52.5072111,13.1449592' 'https://localhost:9999/scrape'
	'''
	connection_closed = False

	def initialize(self, browser_factory):
		self.browser_factory = browser_factory

	def on_connection_close(self):
		self.connection_closed = True

	@gen.coroutine
	def post(self):
		location = self.get_argument('location').split(',')
		if len(location) != 2:
			raise tornado.web.HTTPError(400, 'invalid location parameter')
		lat = float(location[0])
		lng = float(location[1])

		browser = self.browser_factory()

		try:
			scrape = scraper.Scraper(browser, lat, lng)

			while True:
				if self.connection_closed:
					return

				for result in scrape.get_results():
					self.write(result + "\n")
				self.flush()

				if scrape.get_remaining() == 0:
					break

				yield async_sleep(5)
		finally:
			browser.close()

		self.finish()

@gen.coroutine
def async_sleep(seconds):
		yield gen.Task(IOLoop.instance().add_timeout, time.time() + seconds)

def main():
	import argparse
	from selenium import webdriver

	parser = argparse.ArgumentParser(description='Scrape Google Maps distance metrics.')
	parser.add_argument('--port', default=9999, help='the port where we listen for HTTP connections')
	parser.add_argument('--selenium_addr', default='http://localhost:4444/wd/hub',
											help='web address for the selenium webdriver')

	args = parser.parse_args()

	def new_browser():
		return webdriver.Remote(
			desired_capabilities=webdriver.DesiredCapabilities.FIREFOX,
			command_executor=args.selenium_addr)

	application = tornado.web.Application([
		(r"/scrape", ScrapeHandler, dict(browser_factory=new_browser)),
	])

	print 'listening at %d' % args.port
	application.listen(args.port)
	IOLoop.instance().start()

if __name__ == '__main__':
	main()
