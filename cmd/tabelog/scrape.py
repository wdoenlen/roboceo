import json
import requests
import time
import urllib

from bs4 import BeautifulSoup

def get_results(bounds, page):
	'''
	Scrapes the restaurant map endpoint at tabelog.com
	for information about restaurants within the given bounding
	box. Page, an integer from 1 to the number of pages,
	specifies which page of results to return.
	'''
	lat0, lng0, lat1, lng1 = bounds
	params = {
		"maxLat": lat1,
		"minLat": lat0,
		"maxLon": lng1,
		"minLon": lng0,
		"cat0": 0,
		"cat1": '',
		"LstRev": 0,
		"ChkCoupon": 0,
		"pg": page,
		"lst": 100,
		"memo": 0,
		"sw": '',
		"LstSitu": '',
		"LstCosT": 0,
		"LstCos": 0,
		"lunch_flg": 0,
		"SrtT": 'rt',
		"RdoCosTp": 2,
	}

	url = "http://tabelog.com/xml/rstmap"
	url += '?' + urllib.urlencode(params)

	resp = requests.get(url)
	if resp.status_code != 200:
		raise StandardError('bad response %d' % resp.status_code)

	doc = BeautifulSoup(resp.text, 'lxml')

	total_results = int(doc.find("srchinfo").attrs['cnt'])

	markers = doc.find_all("marker")
	results = [marker.attrs for marker in markers]

	return results, total_results

def get_all_results_progress(bounds):
	'''
	Call get_results for each page in the result list. Yield the
	results so far as progress is made.
	'''
	all_results = []
	i = 1
	while True:
		page, total_results = get_results(bounds, i)
		if len(all_results) >= total_results:
			break
		
		all_results += page
		yield all_results

		time.sleep(1)

if __name__ == '__main__':
	bounds = [40.720201, 140.617447, 40.921814, 140.925064]

	for results in get_all_results_progress(bounds):
		print len(results)

		js = json.dumps(results, indent=2, ensure_ascii=False)
		with open('results.json', 'w') as f:
			f.write(js.encode('utf-8'))

	print 'DONE'
