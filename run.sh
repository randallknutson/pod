#!/bin/bash
while ! ./pod $@
do
  sleep 1
  echo "Restarting program..."
done
