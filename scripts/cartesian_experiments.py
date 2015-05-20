#!/usr/bin/python
"""Cartesian execution of options for experiments"""

import itertools
from pprint import pprint
import os

# GROUPS = [
#     ('train', {'type': 'option',
#               'order': 0,
#               'values': ['train5k']}),
#     ('lang', {'type': 'option',
#               'order': 1,
#               'values': 'hungarian,basque,french,korean,polish,swedish'.split(',')}),
#     ('infuse', {'type': 'option',
#               'order': 2,
#               'values': ['true', 'false']}),
#     ('maxmsr', {'type': 'option',
#               'order': 3,
#               'values': '1'.split(',')})
# ]
#
GROUPS = [
    ('train', {'type': 'option',
              'order': 0,
              'values': ['train', 'train5k']}),
    ('lang', {'type': 'option',
              'order': 1,
              'values': 'hungarian,basque,french,korean,polish,swedish'.split(',')}),
    ('infuse', {'type': 'option',
              'order': 2,
              'values': ['true', 'false']}),
    ('maxmsr', {'type': 'option',
              'order': 3,
              'values': '1,2,4,8'.split(',')})
]

# GROUPS = [
#     ('gram', {'type': 'file',
#               'use': 'agg',
#               'order': 0,
#               'values': ['unigram', 'bigram', 'trigram', 'nextunigram', 'nextbigram', 'nexttrigram']}),
#     # ('prev', {'type': 'file',
#     #           'use': 'optional',
#     #           'value': 'prev'}),
#     ('pop', {'type': 'option',
#              'use': 'optional',
#              'value': '-pop'})
# ]

# BASE = """nohup ./chukuparser md -f $conf -td corpus/train4k.hebtb.gold.lattices -tl corpus/train4k.hebtb.pred.lattices -in corpus/dev.hebtb.gold.conll.pred.lattices -ing corpus/dev.hebtb.gold.conll.gold.lattices -om devo.$exp.b32.hebtb.mapping -it 1 -b 32 -p Funcs_Main_POS_Both_Prop -wb -bconc $flags > runstatus.$exp.b32"""
MALEARN = """nohup ./yap malearn -lattice spmrl/train.$lang.gold.conll.tobeparsed.tagged.lattices -raw spmrl/train.$lang.gold.conll.tobeparsed.raw -out $lang.json > malearn.$exp.out"""
MATRAIN = """nohup ./yap ma -dict $lang.json -raw spmrl/$train.$lang.gold.conll.tobeparsed.raw -out $train.$lang.$maxmsr.analyzed.lattices -maxmsrperpos $maxmsr > matrain.$exp.out"""
MADEV = """nohup ./yap ma -dict $lang.json -raw spmrl/dev.$lang.gold.conll.tobeparsed.raw -out dev.$lang.$maxmsr.analyzed.lattices -maxmsrperpos $maxmsr > madev.$exp.out"""
MD = """nohup ./yap md -f conf/standalone.md.yaml -td spmrl/$train.$lang.gold.conll.tobeparsed.tagged.lattices -tl $train.$lang.$maxmsr.analyzed.lattices -in dev.$lang.$maxmsr.analyzed.lattices -ing spmrl/dev.$lang.gold.conll.tobeparsed.tagged.lattices -om devo.$train_$lang_$maxmsr_$infuse.mapping -infusedev=$infuse -it 1 -b 32 -p Funcs_Main_POS_Both_Prop -bconc -pop > runstatus.$exp.out"""

cmds = [MALEARN, MATRAIN, MADEV, MD]
REPLACE_STR = '$exp'

CONF_FILE = 'standalone.md.%s.yaml'

BASE_FILE = 'standalone.base.md.yaml'

# first transform optional to empty, existing
for (name, conf) in GROUPS:
    if conf.get('use', None) == 'optional':
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
    options = {}
    # for i, param in enumerate(execution):
    #     conf_name, conf = GROUPS[i]
    #     # print "\tAt conf %s" % conf_name
    #     # pprint(conf)
    #     # print "\tparam is %s" % str(param)
    #     if conf['type'] == 'option' and param:
    #         print "\t\tadd %s=%s to command line" % (conf_name, str(param))
    #         options[conf_name] = param
    #         # print "\t\tadd %s to command line" % str(conf['value'])
    #         # command_line_options.append(conf['value'])
    #     if conf.get('use', None) == 'optional':
    #         exp_strings.append(conf_name if param else 'no%s' % conf_name)
    #     else:
    #         exp_strings.append(param)
    #     if conf['type'] == 'file':
    #         if conf['use'] == 'agg':
    #             files += conf['values'][:conf['values'].index(param)+1]
    #         if conf['use'] == 'optional' and param:
    #             files.append(param)
    for cmd in cmds:
        execcmd = cmd[:]
        for name, value in zip(map(lambda (k,v):k, GROUPS), execution):
            execcmd = execcmd.replace('$'+name, value)
        execcmd = execcmd.replace('$exp', '_'.join(execution))
        print execcmd
        os.system(execcmd)
    # exp_string = '_'.join(exp_strings)
    # outname = CONF_FILE % exp_string
    # print command_line_options
    # gen_agg_file(files, outname)
    # new_command = BASE.replace('$conf', outname).replace('$exp', exp_string, 2).replace('$flags', ' '.join(command_line_options))
    # print 'Executing %s' % new_command
    # os.system(new_command)
