#!/usr/bin/python
"""Fix truegold hebtb lattices with token indices using pred gold path"""

import depio
from operator import itemgetter
from collections import defaultdict
from pprint import pprint


def run():
    """bla"""
    old = 'train5k.hebtb.gold.lattices'
    new = 'train5k.hebtb.truegold.lattices'

    osents = list(depio.depread(old))
    nsents = list(depio.depread(new))

    zipped = zip(osents, nsents)

    outfile = open('train5k.hebtb.truegold_fixed.lattices', 'w')
    fixtypes = defaultdict(int)

    def fixsimple(osent, nsent):
        """Fix simple"""
        zosent, znsent = zip(*osent), zip(*nsent)
        znsent[-1] = zosent[-1]
        nsent = zip(*znsent)
        return nsent

    log = True

    def matchmiss(osent, nsent):
        j = 0
        i = 0
        numchanges = 0
        while i < len(nsent):
            truemorph = nsent[i]
            predmorph = osent[j]
            if log:
                print '\tAt %s and %s' % (predmorph[2], truemorph[2])
            if predmorph[2] == truemorph[2]:
                if log:
                    print '\t\tFixing1 %s with %s' % (truemorph[2], predmorph[2])
                truemorph[-1] = predmorph[-1]
                j += 1
                i += 1
                numchanges += 1
            elif j < len(osent)-1 and ''.join([predmorph[2], osent[j+1][2]]) == truemorph[2]:
                if log:
                    print '\t\tFixing2 %s with %s' % (truemorph[2], predmorph[2])
                truemorph[-1] = predmorph[-1]
                j += 2
                i += 1
                numchanges += 1
            elif j < len(osent)-1 and ''.join([predmorph[2], osent[j+1][2][1:]]) == truemorph[2]:
                if log:
                    print '\t\tFixing2 %s with %s' % (truemorph[2], predmorph[2])
                truemorph[-1] = predmorph[-1]
                j += 2
                i += 1
                numchanges += 1
            elif i < len(nsent)-1 and ''.join([truemorph[2], nsent[i+1][2]]) == predmorph[2]:
                if log:
                    print '\t\tFixing3 %s with %s' % (truemorph[2], predmorph[2])
                truemorph[-1] = predmorph[-1]
                i += 1
                numchanges += 1
            elif i > 0 and ''.join([nsent[i-1][2], truemorph[2]]) == predmorph[2]:
                if log:
                    print '\t\tFixing4 %s with %s' % (truemorph[2], predmorph[2])
                truemorph[-1] = predmorph[-1]
                j += 1 
                i += 1
                numchanges += 1
            elif truemorph[2][:3] == predmorph[2][:3] and len(osent)>j+1 and \
                    osent[j+1][4] == 'S_PRN':
                if log:
                    print '\t\tFixing6 %s with %s' % (truemorph[2], predmorph[2])
                truemorph[-1] = predmorph[-1]
                j += 2
                i += 1
                numchanges += 1
            elif truemorph[2][:3] == predmorph[2][:3] and len(nsent)>i+1 and \
                    nsent[i+1][4] == 'S_PRN':
                if log:
                    print '\t\tFixing8 %s with %s' % (truemorph[2], predmorph[2])
                    print '\t\tFixing8 %s with %s' % (nsent[i+1][2], predmorph[2])
                truemorph[-1] = predmorph[-1]
                nsent[i+1][-1] = predmorph[-1]
                j += 1
                i += 2
                numchanges += 2
            elif len(nsent)> i+1 and len(osent) > j+1 and \
            ''.join([truemorph[2], nsent[i+1][2]]) == ''.join([predmorph[2], osent[j+1][2]]):
                if log:
                    print '\t\tFixing10 %s with %s' % (truemorph[2], predmorph[2])
                    print '\t\tFixing10 %s with %s' % (nsent[i+1][2], osent[j+1][2])
                truemorph[-1] = predmorph[-1]
                nsent[i+1][-1] = osent[j+1][-1]
                j += 2
                i += 2
                numchanges += 2
            elif set([truemorph[2][:3], predmorph[2][:3]]) == set(['EM', 'AT']) and \
                      len(osent)>j+1 and len(nsent)>i+1 and \
                    nsent[i+1][4] == 'S_PRN' and osent[j+1][4] == 'S_PRN':
                if log:
                    print '\t\tFixing9 %s with %s' % (truemorph[2], predmorph[2])
                truemorph[-1] = predmorph[-1]
                j += 1
                i += 1
                numchanges += 1
            elif set([truemorph[2], predmorph[2]]) == set(['ATH', 'AT']) and \
                    truemorph[4] == 'S_PRN' and truemorph[4] == 'S_PRN':
                if log:
                    print '\t\tFixing13 %s with %s' % (truemorph[2], predmorph[2])
                truemorph[-1] = predmorph[-1]
                j += 1
                i += 1
                numchanges += 1
            elif truemorph[4] == 'IN' and len(nsent) > i+1 and nsent[i+1][4] == 'S_PRN':
                if log:
                    print '\t\tFixing7 %s with %s' % (truemorph[2], predmorph[2])
                    print '\t\tFixing7 %s with %s' % (nsent[i+1][-1], predmorph[2])
                truemorph[-1] = predmorph[-1]
                nsent[i+1][-1] = predmorph[-1]
                j += 1
                i += 2
                numchanges += 2
            elif truemorph[2] == 'B' and truemorph[4] == 'PREPOSITION' and \
                 len(nsent) > i+1 and len(osent) > j+1 and \
                 predmorph[4] == 'IN' and osent[j+1][4] == 'S_PRN':
                if log:
                    print '\t\tFixing7 %s with %s' % (truemorph[2], predmorph[2])
                    print '\t\tFixing7 %s with %s' % (nsent[i+1][-1], predmorph[2])
                truemorph[-1] = predmorph[-1]
                nsent[i+1][-1] = predmorph[-1]
                j += 2
                i += 2
                numchanges += 2
            elif len(nsent) > i+2 and predmorph[2] == ''.join([truemorph[2], nsent[i+1][2], nsent[i+2][2]]):
                if log:
                    print '\t\tFixing12 %s with %s' % (truemorph[2], predmorph[2])
                    print '\t\tFixing12 %s with %s' % (nsent[i+1][-1], predmorph[2])
                    print '\t\tFixing12 %s with %s' % (nsent[i+2][-1], predmorph[2])
                truemorph[-1] = predmorph[-1]
                j += 1
                i += 3
                numchanges += 3
            elif len(nsent) > i+2 and nsent[i+1][4] == 'IN' and nsent[i+2][4] == 'S_PRN':
                if log:
                    print '\t\tFixing11 %s with %s' % (truemorph[2], predmorph[2])
                    print '\t\tFixing11 %s with %s' % (nsent[i+1][-1], predmorph[2])
                    print '\t\tFixing11 %s with %s' % (nsent[i+2][-1], predmorph[2])
                truemorph[-1] = predmorph[-1]
                nsent[i+1][-1] = predmorph[-1]
                nsent[i+2][-1] = predmorph[-1]
                j += 1
                i += 3
                numchanges += 3
            elif len(nsent) > i+2 and ''.join([truemorph[2], nsent[i+2][2]]) == predmorph[2]:
                if log:
                    print '\t\tFixing7 %s with %s' % (truemorph[2], predmorph[2])
                    print '\t\tFixing7 %s with %s' % (nsent[i+1][-1], predmorph[2])
                    print '\t\tFixing7 %s with %s' % (nsent[i+2][-1], predmorph[2])
                truemorph[-1] = predmorph[-1]
                nsent[i+1][-1] = predmorph[-1]
                nsent[i+2][-1] = predmorph[-1]
                j += 1
                i += 3
                numchanges += 3
            elif truemorph[2] == 'H' and i>0 and len(nsent)>i+1 and len(osent)>j+1 and \
            ''.join([osent[j-1][2], predmorph[2]]) == ''.join([nsent[i-1][2], nsent[i+1][2]]):
                if log:
                    print '\t\tFixing5 %s with %s' % (truemorph[2], predmorph[2])
                i += 1
                truemorph[-1] = predmorph[-1]
                numchanges += 1
            else:
                i += 1
        return nsent, numchanges == len(nsent)

    for num, (osent, nsent) in enumerate(zipped):
        print 'At %s' % str(num)
        fget = itemgetter(2)
        oforms, nforms = map(fget, osent), map(fget, nsent)
        out = nsent
        success = False
        if len(osent) == len(nsent) and oforms == nforms:
            out = fixsimple(osent, nsent)
            fixtypes['proper'] += 1
            success = True
        else:
            out, success = matchmiss(osent, nsent)
            fixtypes['match' if success else 'nomatch'] += 1
        if not success:
            print 'Failed at %s' % str(num)
        outfile.write(depio.depstring(out))
    pprint(fixtypes)
    print 'Total %s' % str(sum(fixtypes.values()))

run()
