yap - Yet Another Parser
===========

yap is yet another parser written in Go. It was implemented to test the
hypothesis of my MSc thesis on Joint Morpho-Syntactic Processing of MRLs in a
Transition Based Framework at IDC Herzliya with my advisor, Reut Tsarfaty.
A paper on the morphological analysis and disambiguation aspect for Modern Hebrew
and Universal Dependencies was accepted to COLING 2016.

yap is currently provided with a model for Modern Hebrew, trained on a heavily updated
version of the SPMRL 2014 Hebrew treebank. We hope to publish the updated
treebank soon as well.

yap contains an implementation of the framework and parser of zpar from Z&N 2011 ([Transition-based Dependency Parsing with Rich Non-local Features by Zhang and Nivre, 2011](http://www.aclweb.org/anthology/P11-2033.pdf)) with flags for precise output parity (i.e. bug replication).

yap is under active development and documentation.

***DO NOT USE FOR PRODUCTION***

Requirements
-----------
- [http://www.golang.org](Go)
- bzip2
- 4-16 CPU cores
- ~4.5GB RAM 

Compilation
-----------
- Download and install Go
- Setup a Go environment:
    - Create a directory (usually per workspace/project) ``mkdir yapproj; cd yapproj``
    - Set ``$GOPATH`` environment variable to your workspace: ``export GOPATH=path/to/yapproj ``
    - In the workspace directory create 3 subdirectories: ``mkdir src pkg bin``
    - cd into the src directory ``cd src``
- Clone the repository in the src folder of the workspace, then:
```
go get .
go build .
./yap
```
- Unzip the Hebrew MD model: ``bunzip2 data/hebmd.b32.gz``

You may want to use a go workspace manager or have a shell script to set ``$GOPATH`` to <.../yapproj>

Processing Modern Hebrew
-----------
Currently only Morphological Analysis and Disambiguation of pre-tokenized Hebrew
text is supported. For Hebrew Morphological Analysis, the input format should
have tokens separated by a newline, with another newline to separate sentences.
The lattice format as output by the analyzer can be used as-is for
disambiguation.

For example:
```
עשרות
אנשים
מגיעים
מתאילנד
...

כך
אמר
ח"כ
...
```

Commands for morphological analysis and disambiguation:

```
./yap hebma -prefix data/bgulex/bgupreflex_withdef.utf8.hr -lexicon data/bgulex/bgulex.utf8.hr -raw input.raw -out lattices.conll
./yap md -m data/hebmd -f conf/standalone.md.yaml -in lattices.conll -om output.conll
```

Citation
-----------
If you make use of this software for research, we would appreciate the following citation:
```
@InProceedings{moretsarfatycoling2016,
  author = {Amir More and Reut Tsarfaty},
  title = {Data-Driven Morphological Analysis and Disambiguation for Morphologically Rich Languages and Universal Dependencies},
  booktitle = {Proceedings of COLING 2016},
  year = {2016},
  month = {december},
  location = {Osaka}
}
```

HEBLEX, a Morphological Analyzer for Modern Hebrew in yap, relies on a slightly modified version of the BGU Lexicon. Please acknowledge and cite the work on the BGU Lexicon with this citation:
```
@inproceedings{adler06,
    Author = {Adler, Meni and Elhadad, Michael},
    Booktitle = {ACL},
    Crossref = {conf/acl/2006},
    Editor = {Calzolari, Nicoletta and Cardie, Claire and Isabelle, Pierre},
    Ee = {http://aclweb.org/anthology/P06-1084},
    Interhash = {6e302df82f4d7776cc487d5b8623d3db},
    Intrahash = {c7ac3ecfe40d039cd6c9ec855cb432db},
    Keywords = {dblp},
    Publisher = {The Association for Computer Linguistics},
    Timestamp = {2013-08-13T15:11:00.000+0200},
    Title = {An Unsupervised Morpheme-Based HMM for {H}ebrew Morphological
        Disambiguation},
    Url = {http://dblp.uni-trier.de/db/conf/acl/acl2006.html#AdlerE06},
    Year = 2006,
    Bdsk-Url-1 = {http://dblp.uni-trier.de/db/conf/acl/acl2006.html#AdlerE06}}
```

License
-----------
This software is released under the terms of the [https://www.apache.org/licenses/LICENSE-2.0](Apache License, Version 2.0).


Contact
-----------
You may contact me at mygithubuser at gmail or Reut Tsarfaty at reutts at openu dot ac dot il
