#!/bin/bash

URL="http://192.167.137.106:8080/invoke/hello"

echo "Sending 50 parallel POST requests to $URL ..."
for i in {1..50}
do
  {
    echo "----- Request $i -----"
    curl -X POST "$URL"
    echo -e "\n----------------------\n"
  } &
done

wait
echo "All 50 parallel requests completed."
