#!/bin/sh -l

/app

time=$(date)
echo "time=$time" >> $GITHUB_OUTPUT

echo $GITHUB_OUTPUT
