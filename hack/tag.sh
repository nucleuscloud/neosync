#!/bin/sh
git fetch
git tag "$1" origin/main
git push origin "$1"
