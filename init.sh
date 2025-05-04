#!/bin/sh

if ! (type task >/dev/null 2>&1); then
  echo "Install \"task\" and rerun this script."
  echo "https://taskfile.dev/ja-JP/installation/"
fi

task init
