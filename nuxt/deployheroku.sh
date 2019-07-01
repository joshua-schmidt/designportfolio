#!/usr/bin/env bash

# abort on errors
set -e

yarn precommit
yarn email:build

sed -i 's/.env/dummy/g' .gitignore

git init
git remote add heroku https://git.heroku.com/annettewebsite.git
git add -A
git commit -m "deploying to heroku"
git push heroku master -f

sed -i 's/dummy/.env/g' .gitignore

rm -rf .git
