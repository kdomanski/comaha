#!/usr/bin/env bash

set -e
set -u

urlencode_sed="$(echo cy8lLyUyNS9nCnMvIC8lMjAvZwpzLyAvJTA5L2cKcy8hLyUyMS9nCnMvIi8lMjIvZwpzLyMvJTIzL2cKcy9cJC8lMjQvZwpzL1wmLyUyNi9nCnMvJ1wnJy8lMjcvZwpzLygvJTI4L2cKcy8pLyUyOS9nCnMvXCovJTJhL2cKcy8rLyUyYi9nCnMvLC8lMmMvZwpzLy0vJTJkL2cKcy9cLi8lMmUvZwpzL1wvLyUyZi9nCnMvOi8lM2EvZwpzLzsvJTNiL2cKcy8vJTNlL2cKcy8/LyUzZi9nCnMvQC8lNDAvZwpzL1xbLyU1Yi9nCnMvXFwvJTVjL2cKcy9cXS8lNWQvZwpzL1xeLyU1ZS9nCnMvXy8lNWYvZwpzL2AvJTYwL2cKcy97LyU3Yi9nCnMvfC8lN2MvZwpzL30vJTdkL2cKcy9+LyU3ZS9nCnMvICAgICAgLyUwOS9nCg== | base64 -d)"
file="$1"

function urlencode() {
  echo "$1" | sed -f <(echo "$urlencode_sed")
}

sha1=$(sha1sum $file | cut -f1 -d ' ' | xxd -r -p | base64)
sha256=$(sha256sum $file | cut -f1 -d ' ' | xxd -r -p | base64)

url="127.0.0.1:8090"

echo "$sha1"
curl -i -XPOST "$url/admin/add_payload?sha1=$(urlencode $sha1)&sha256=$(urlencode $sha256)&version=$(urlencode $2)&channel=$3" --data-binary "@$1"
