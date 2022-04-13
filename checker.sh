#!/bin/bash
for file in $(ls -a | grep gz);
do key=$(zcat $file | tail -n 1 | awk '{print $1, ":", $2}'|tr -d " ");
echo "key: $key";
echo "get $key" | nc -N 127.0.0.1 11211;
done