#!/bin/bash

set -e

cat google-10000-english.txt | \
	perl -MList::Util=shuffle -e 'print shuffle<STDIN>' | \
	parallel -j 4 --pipe --line-buffer python get_slide_urls.py | \
	parallel -j 8 wget -q -P slides
