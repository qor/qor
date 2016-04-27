#!/usr/bin/env bash

root=$GOPATH/src/github.com/qor

tmp_dir=$(mktemp -d)

trap "rm -rf $tmp_dir" EXIT

function pull_chances {
    local pkg=$(basename $1)
    local pkg_path=$root/$1

    if [ ! -d "$pkg_path/.git" ]; then
        echo "$pkg not a git repo. skipped."
        return 0
    fi

    cd $pkg_path

    if [[ `git status -s --untracked-files=no` != "" ]]; then
        echo "$pkg is not clean. please stash or commit your changes."
        exit 1
    fi

    git checkout master >> /dev/null 2>&1 || {
        echo "failed to update $pkg"
        touch $tmp_dir/failed
        exit 1
    }

    (git pull --rebase --quiet && echo -e "\033[31mUpdating $pkg...\033[0m") || {
        echo -e "\033[31mfailed to update $pkg\033[0m"
        touch $tmp_dir/failed
        exit 1
    }
}

function update_git_repo {
    for pkg in $(ls $root); do
        pull_chances $pkg &
    done
}

update_git_repo

wait

if [[ -f "$tmp_dir/failed" ]]; then
    echo "failed"
    exit 1
else
    echo "done"
fi
