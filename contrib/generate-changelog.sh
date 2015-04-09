#!/usr/bin/env bash
# Generates a changelog of all merges from a given release all the way to HEAD.

REPO=https://github.com/kelseyhightower/confd

usage() {
    echo "Usage: $0 <FROM> [TO]"
}

# Print usage summary if user didn't specify a beginning
if [ -z "$1" ];
then
    usage
    exit 1
fi

FROM=$1
TO=${2:-HEAD}

printf "### $TO\n\n"

git --no-pager log --merges --format="%h %b" $FROM..$TO
