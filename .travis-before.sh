#!/usr/bin/env bash

set -e

rm -rf ${GCR_REPO}
git clone https://mritd:${GITHUB_TOKEN}@github.com/mritd/gcr.git ${GCR_REPO}
