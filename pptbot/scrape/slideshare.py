'''This module contains utilities for scraping slideshare.net'''
import requests
from urlparse import urlsplit, urlunsplit
import lxml.html

def search(query):
	'''Get the URLs for the presentations returned on the first page
	of search results on slideshare.net.'''
	url = "http://www.slideshare.net/search/slideshow?q=%s" % query
	resp = requests.get(url)
	if resp.status_code != 200:
		raise StandardError("search: bad response %d" % resp.status_code)
	doc = lxml.html.fromstring(resp.text)
	results = []
	for link in doc.cssselect('a.title-link'):
		href = link.get('href')
		path = urlsplit(href)[2]
		without_query = urlunsplit(('http', 'www.slideshare.net', path, '', ''))
		results.append(without_query)
	return results

def get_slide_urls(preso_url):
	'''Parses the given slideshare.net presentation page and returns
	the urls for the presentation's high-resolution slide previews.'''
	resp = requests.get(preso_url)
	if resp.status_code != 200:
		raise StandardError("get_slide_urls: bad response %d" % resp.status_code)
	doc = lxml.html.fromstring(resp.text)
	results = []
	for img in doc.cssselect('img.slide_image'):
		img_href = img.get('data-full')
		href_parts = urlsplit(img_href)
		without_query = urlunsplit(href_parts[0:3] + ('', ''))
		results.append(without_query)
	return results
