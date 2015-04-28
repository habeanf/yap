#!/usr/bin/python
"""Cartesian execution of options for experiments"""

import itertools
from pprint import pprint
import os

GROUPS = [
    ('gram', {'type': 'file',
              'use': 'agg',
              'order': 0,
              'values': ['unigram', 'bigram', 'trigram', 'nextunigram', 'nextbigram', 'nexttrigram']}),
    # ('prev', {'type': 'file',
    #           'use': 'optional',
    #           'value': 'prev'}),
    ('pop', {'type': 'option',
             'use': 'optional',
             'value': '-pop'})
]

BASE = """nohup ./chukuparser md -f $conf -td corpus/train4k.hebtb.gold.lattices -tl corpus/train4k.hebtb.pred.lattices -in corpus/dev.hebtb.gold.conll.pred.lattices -ing corpus/dev.hebtb.gold.conll.gold.lattices -om devo.$exp.b32.hebtb.mapping -it 1 -b 32 -p Funcs_Main_POS_Both_Prop -wb -bconc $flags > runstatus.$exp.b32"""

REPLACE_STR = '$exp'

CONF_FILE = 'standalone.md.%s.yaml'

BASE_FILE = 'standalone.base.md.yaml'

# first transform optional to empty, existing
for (name, conf) in GROUPS:
    if conf['use'] == 'optional':
        conf['values'] = [None, conf['value']]

conf_values = map(lambda (name, conf): conf['values'], GROUPS)

executions = list(itertools.product(*conf_values))

def gen_agg_file(values, out_name):
    with open(out_name, 'w') as outf:
        for value in values:
            with open(value) as inf:
                outf.write(inf.read())

for execution in executions:
    print 'At execution %s' % str(execution)
    files = [BASE_FILE]
    exp_strings = []
    command_line_options = []
    for i, param in enumerate(execution):
        conf_name, conf = GROUPS[i]
        # print "\tAt conf %s" % conf_name
        # pprint(conf)
        # print "\tparam is %s" % str(param)
        if conf['type'] == 'option' and param:
            # print "\t\tadd %s to command line" % str(conf['value'])
            command_line_options.append(conf['value'])
        if conf['use'] == 'optional':
            exp_strings.append(conf_name if param else 'no%s' % conf_name)
        else:
            exp_strings.append(param)
        if conf['type'] == 'file':
            if conf['use'] == 'agg':
                files += conf['values'][:conf['values'].index(param)+1]
            if conf['use'] == 'optional' and param:
                files.append(param)

    exp_string = '_'.join(exp_strings)
    outname = CONF_FILE % exp_string
    gen_agg_file(files, outname)
    new_command = BASE.replace('$conf', outname).replace('$exp', exp_string, 2).replace('$flags', ' '.join(command_line_options))
    print 'Executing %s' % new_command
    os.system(new_command)
