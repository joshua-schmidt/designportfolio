#!/usr/bin/env bash

# abort on errors
set -e

yarn build

sed -i 's/config/dummy/g' .gitignore

git init
git remote add heroku https://git.heroku.com/annettewebsiteamp.git
git add -A
git commit -m "deploying to heroku"
git push heroku master -f

sed -i 's/dummy/config/g' .gitignore

rm -rf .git
