function update_git_repo {
  for path in $( find $1 -maxdepth 2 -type d ); do
    cd $path
    if test -n "$(find . -maxdepth 1 -name '.git' -print -quit)"; then
      echo "\033[31mUpdating $(basename $path)...\033[0m"
      git pull
    fi
  done
}

update_git_repo $GOPATH/src/github.com/qor
