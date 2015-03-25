example: cd example && go get && RUNNER_EXTRA_DIRS="$GOPATH/src/github.com/qor,$GOPATH/src/github.com/jinzhu" fresh
scss:    for n in $(find -iname '*.scss'); do scss --sourcemap=none --watch $n:$(echo $n | sed 's/\.scss/.css/' | sed 's/\/scss\//\//') &; done
coffee:  for n in $(find -iname '*.coffee'); do coffee --watch $n -o $(echo $n | sed 's/\.coffee/.js/' | sed 's/\/coffee\//\//') &; done
tests:   goconvey --port=9999 -cover=true -depth=-1
