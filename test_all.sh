function gotest_path {
  for path in $( find $1 -maxdepth 2 -type d ); do
    cd $path
    if test -n "$(find . -maxdepth 1 -name '*test.go' -print -quit)"; then
      echo "($path)"
      echo "\033[31mTesting $(basename $path) with mysql...\033[0m"
      TEST_DB=mysql go test
      echo "\033[31mTesting $(basename $path) with postgres...\033[0m"
      TEST_DB=postgres go test
    fi
  done
}

gotest_path $GOPATH/src/github.com/qor
