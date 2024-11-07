#!/bin/sh
air &
cd "$(dirname "$0")/../web" && npm install && npm run watch:css
# cd "$(dirname "$0")/../web" && npm install && npm run watch:css--loglevel=verbose
# wait