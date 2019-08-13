#!/bin/bash

# abort on errors
set -e

changes() {
  git diff --stat --cached -- graphql/
}

travis_ignore="[skip ci]"

if ! changes | grep -E "graphql/" ; then
  echo "no ci changes found"
  sed -i.bak -e "1s/^/$travis_ignore /" "$GIT_DIR/COMMIT_EDITMSG"
else
  echo "ci changes found"
fi
