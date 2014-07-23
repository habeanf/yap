#!/usr/bin/python

import sys

sents = int(sys.argv[2])
fnames = [sys.argv[1]]
for fname in fnames:
	lines = open(fname).readlines()
	i=0
	out = open("%d.%s" % (sents,fname),'w')
	for l in lines:
		if len(l)<2:
			i+=1
		out.write(l)
		if i>=sents:
			break
	out.close()
