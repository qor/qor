#!/bin/bash

function migrate {
  for package in $( find $1 -maxdepth 1 -type d ); do
    echo "migrating $(basename $package)..."

    for file in $( grep "github.com/qor/qor/$(basename $package)" -l -r . ); do
      if [[ $file == *".go" ]]
      then
        echo "updating package $(basename $package)'s import path for file $file "
        sed -i '' s/\\/qor\\/qor\\/$(basename $package)/\\/qor\\/$(basename $package)/g $file
      fi
    done
  done
}

migrate $GOPATH/src/github.com/qor
