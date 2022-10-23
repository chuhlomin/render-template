#!/bin/sh -l

echo "Template: $INPUT_TEMPLATE"

echo "result_from_entrypoint=123" >> $GITHUB_OUTPUT

/app
