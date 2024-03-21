#!/usr/bin/bash

procedure=$1
dirname=$(dirname "$0")

if [ "$procedure" == "move" ]; then
  sites_order=("site-2" "site-1")
elif [ "$procedure" == "stop" ]; then
  sites_order=("site-1" "site-2")
else
  echo "Wrong procedure... only move and stop are supported"
  exit 1
fi

start_time=$(date +%s%N)
for _ in {1..5}; do
  for site in "${sites_order[@]}"; do
    echo "Start procedure $procedure for site $site..."
    python3 smclient.py -c $dirname/sm-client-config.yaml -v $procedure $site
    if [ $? -ne 0 ]; then
      echo "Procedure fails, exit"
      exit 1
    fi
  done
done
end_time=$(date +%s%N)

elapsed=$(($((end_time - start_time)) / 10))
echo "Elapsed time is: $elapsed nanoseconds"
