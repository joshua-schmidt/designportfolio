#!/bin/bash

# abort on errors
set -e

yarn upgrade
cd amp
yarn upgrade
cd ../docs
yarn upgrade
cd ../init
yarn upgrade
cd ../nuxt
yarn upgrade
cd ../shortlink
yarn upgrade
cd ../graphql
go get -u -m
cd ..
