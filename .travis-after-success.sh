#!/usr/bin/env bash

set -e

if [ -f "imgsync_report" ]; then
  openssl aes-256-cbc -K encrypted_39d2a83529c0_key -iv encrypted_39d2a83529c0_iv -in msgsend.yaml.enc -out msgsend.yaml -d
  echo "${TRAVIS_JOB_NAME} success!" >> report
  cat imgsync_report >> report
  msgsend --config msgsend.yaml txt --file report
fi

export TZ=UTC-8
cd ${GCR_REPO}
if [ -n "$(git status --porcelain)" ]; then
  git add .
  git commit -m "Travis CI Auto Update(`date +'%Y-%m-%d %H:%M:%S'`)"
  git push https://mritd:${GITHUB_TOKEN}@github.com/mritd/gcr.git
fi
