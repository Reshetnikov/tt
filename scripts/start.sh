#!/bin/sh
air &
cd "$(dirname "$0")/../web" && npm run watch:css
wait