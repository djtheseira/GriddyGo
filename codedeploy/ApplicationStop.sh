#!/bin/bash

if pgrep -x "griddy" > /dev/null
then 
  echo "Running"
  sudo killall --signal SIGINT griddy
else 
  echo "Sleeping"
fi