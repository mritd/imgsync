#!/usr/bin/env bash

set -e

export TZ=UTC-8
cd ${HOME}/manifests
if [ ! git diff-index --quiet HEAD ]; then
  git add .
  git commit -m "Travis CI Auto Update(`date +'%Y-%m-%d %H:%M:%S'`)"
  git push https://mritd:${GITHUB_TOKEN}@github.com/mritd/gcr.git
fi
