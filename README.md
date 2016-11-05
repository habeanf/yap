yap - Yet Another Parser
===========

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
    - Set ``$GOPATH`` environment variable to your workspace: ``export GOPATH=<path/to>/yapproj ``
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
./yap hebma -prefix databgulex/bgupreflex_withdef.utf8.hr -lexicon data/bgulex/bgulex.utf8.hr -raw input.raw -out lattices.conll
./yap md -m data/hebmd -f conf/standalone.md.yaml -in lattices.conll -om output.conll -bconc -nolemma
```

Citation
-----------
If you make use of this software for research, we would appreciate the following citation:
```
@inproceedings{2016COLINGMoreTsarfaty,
  author = {{More}, A. and {Tsarfaty}, R.},
  ...
}
```

License
-----------
This software is released under the terms of the [https://www.apache.org/licenses/LICENSE-2.0](Apache License, Version 2.0).
