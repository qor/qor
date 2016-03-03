function migrate {
  for package in $( find $1 -maxdepth 1 -type d ); do
    for file in $( grep "github.com/qor/qor/$(basename $package)" -r **/*.go -l ); do
      echo "updating package import path $(basename $package) for file $file "
      sed -i '' s/\\/qor\\/qor\\/$(basename $package)/\\/qor\\/$(basename $package)/g $file
    done
  done
}

migrate $GOPATH/src/github.com/qor
