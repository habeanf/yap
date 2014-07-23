# -encoding:utf8-
import sys
import depio
import re
import operator
from itertools import chain
from pprint import pprint

morph_funcs = [
   ['seg', [2,-1]],
   ['pos-seg', [2, 5, -1]],
   ['pos-seg-feats', [2, 5, 6, -1]]
]

def get_morph_func(name):
   m = dict(morph_funcs)[name]
   return operator.itemgetter(*m)

def f1(precision, recall):
   return 2*(precision*recall)/(precision+recall)

def precision(true_positives, test_positives):
   return true_positives/test_positives

def recall(true_positives, condition_positives):
   return true_positives/condition_positives

def eval(output, reference):
   for index, word in enumerate(output):
      ref_word = reference[index]
      assert word[0] == ref_word[0]
      if g_reP.match( word[0] ) :
         continue
      if word[2] == ref_word[2]:
         correct_head += 1
         if word[3] == ref_word[3]:
            correct_label += 1
      else:
         total_uem = 0
      total += 1
   return correct_head, correct_label, total, total_uem

def flatten(listOfLists):
    "Flatten one level of nesting"
    return chain.from_iterable(listOfLists)

def get_set_elements(sent, morph_func):
   return zip([sent[0]]*len(sent[1]), map(morph_func, sent[1]))

if __name__ == '__main__':
   # file_output = list(depio.depread(sys.argv[1]))
   # prob_sent_ids = [4, 24, 28, 32, 40, 53, 55, 117, 136, 155, 168, 170, 182, 209, 224, 228, 255, 262, 270, 285, 330, 332, 351, 378, 379, 427, 441, 447, 448, 451, 454, 462, 469, 473, 481, 492, 493]
   # prob_sents = []
   # for sid in prob_sent_ids:
   #    s = file_output[sid]
   #    print 'offset: %d' % sid
   #    for m in s:
   #       print '\t'.join(m)
   #    print ''

   file_output = list(enumerate(depio.depread(sys.argv[1])))
   file_ref = list(enumerate(depio.depread(sys.argv[2])))
   print '\t'.join(['comparison-function', 'gold', 'pred', 'true-positive', 'precision', 'recall', 'f1'])
   for morph_func_name, morph_func_opts in morph_funcs[:1]:
      morph_func = get_morph_func(morph_func_name)
      output_set_el = map(lambda s: get_set_elements(s, morph_func), file_output)
      prob_sent = filter(lambda el:len(el)>len(set(el)), output_set_el)
      d={}
      if len(prob_sent)>0:
         for s in prob_sent:
            t=dict()
            for m in s:
               if m[1] in t:
                  t[m[1]]+=1
               else:
                  t[m[1]] = 1
            for k,v in t.items():
               if v > 1:
                  print (s[0][0], k)
                  if k[0] in d:
                     d[k[0]]+=1
                  else:
                     d[k[0]]=1
            print '\n'
      print 'Total %d prob sents' % len(prob_sent)
      print 'Prob sent nums: %s' % (str(map(lambda s:s[0][0], prob_sent)))
      pprint(d.items())
      output_set = set(flatten(map(lambda s: get_set_elements(s, morph_func), file_output)))
      ref_set = set(flatten(map(lambda s: get_set_elements(s, morph_func), file_ref)))
      pred_segments = len(output_set)
      gold_segments = len(ref_set)
      true_positives = float(len(output_set.intersection(ref_set)))
      p = precision(true_positives, pred_segments)
      r = recall(true_positives, gold_segments)
      f1_score = f1(p, r)
      print '\t'.join(map(str,[morph_func_name,gold_segments,pred_segments,int(true_positives),p,r,f1_score]))
      # print float(correct_head)/total, float(correct_label)/total, float(total_uem)/total_sent