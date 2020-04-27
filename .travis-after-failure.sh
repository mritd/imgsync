#!/usr/bin/env bash

set -e

openssl aes-256-cbc -K $encrypted_39d2a83529c0_key -iv $encrypted_39d2a83529c0_iv -in msgsend.yaml.enc -out msgsend.yaml -d
msgsend --config msgsend.yaml txt "❌ ${TRAVIS_JOB_NAME} failed!"
