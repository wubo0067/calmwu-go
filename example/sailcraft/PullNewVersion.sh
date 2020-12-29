#!/bin/bash

git fetch --all
git reset --hard origin/master
git pull
find . -name "*.sh"|xargs chmod +x