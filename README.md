yap - Yet Another Parser
===========

Compilation
-----------
- [http://www.golang.org](Download and install Go)
- Setup a Go environment:
    - Create a directory (usually per workspace/project) ``mkdir yapproj; cd yap``
    - Set ``$GOPATH`` environment variable do your workspace: ``export GOPATH=`pwd` ``
    - In that directory create 3 subdirectories: ``mkdir src pkg bin``
    - cd into the src directory ``cd src``
- Clone the repository in the src folder of the workspace, then:

```
go get .
go build .
./yap
```
