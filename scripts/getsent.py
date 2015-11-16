#!/usr/bin/python

import sys
import depio

sentnum = int(sys.argv[2])
fnames = [sys.argv[1]]
for fname in fnames:
	sents = list(depio.depread(fname))
	i=0
	out = open("%d.%s" % (sentnum,fname),'w')
	for outl in sents[sentnum]:
		out.write('\t'.join(outl) + '\n')
	out.write('\n')
	out.close()
