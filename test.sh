#!/bin/bash

URL="http://10.88.100.251:8080/invoke/hello"

echo "Sending 50 POST requests to $URL ..."
for i in {1..50}
do
  echo "----- Request $i -----"
  curl -X POST "$URL"
  echo -e "\n----------------------\n"
done

echo "All 50 requests completed."

