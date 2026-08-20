#!/bin/bash

set -euf -o pipefail

echo -n "Version: "
read -r newversion

if [[ ! "$newversion" =~ ^[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.-]+)?$ ]]; then
  echo "Version must be semantic version text such as 0.0.6 or 0.0.6-rc.1"
  exit 1
fi

if git rev-parse --verify --quiet "refs/tags/v$newversion" >/dev/null; then
  echo "Tag: $newversion already exists"
  exit 1
else
  echo "Releasing $newversion"
fi

echo "Update CHANGELOG.md and press enter"
read -r

git add CHANGELOG.md

if ! git diff --cached --quiet -- CHANGELOG.md; then
  git commit -m"Release version $newversion"
fi

git tag "v$newversion" -m "v$newversion"

git push origin main "v$newversion"
