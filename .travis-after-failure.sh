#!/usr/bin/env bash

set -e

openssl aes-256-cbc -K $encrypted_0a6446eb3ae3_key -iv $encrypted_0a6446eb3ae3_iv -in msgsend.yaml.enc -out msgsend.yaml -d
msgsend --config msgsend.yaml txt "❌ ${TRAVIS_JOB_NAME} failed!"
