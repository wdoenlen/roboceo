'''Utility that reads query words from stdin (one per line)
and returns slide URLs from a random presentation on the slideshare
search results page for that word. Pipe it random words to get random
slides.'''

import slideshare
import random
import fileinput

for word in fileinput.input():
	candidates = slideshare.search(word)
	if len(candidates) == 0:
		continue
	preso = random.sample(candidates, 1)[0]
	for url in slideshare.get_slide_urls(preso):
		print url
