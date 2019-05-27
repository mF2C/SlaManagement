#!/usr/bin/env bash
#
set -e
set -u

TYPE=$(git log -n 1 | grep -i "change-type:" | sed -e 's/ *change-type: *\(.*\)/\1/I')
if [ $? -ne 0 ]; then
  echo "Change-type not found on last commit"
  return 1
fi

versionist || { 
  echo "Error running versionist" && exit 
}
VERSION=$(versionist get version)

git add CHANGELOG.md
git commit -m "Release $VERSION"

resources/bin/release.sh $TYPE

