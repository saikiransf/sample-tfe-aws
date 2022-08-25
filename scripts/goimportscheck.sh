#!/usr/bin/env bash

set -euo pipefail

# Check goimports
echo "==> Checking the code complies with goimports requirements..."

# We only require goimports to have been run on files that were changed
# relative to the main branch, so that we can gradually create more consistency
# rather than bulk-changing everything at once.

declare -a target_files
# "readarray" will return an "unbound variable" error if there isn't already
# at least one element in the target array. "readarray" will overwrite this
# item, though.
target_files[0]=""

base_branch="origin/main"

# HACK: If we seem to be running inside a GitHub Actions pull request check
# then we'll use the PR's target branch from this variable instead.
if [[ -n "${GITHUB_BASE_REF:-}" ]]; then
  base_branch="origin/$GITHUB_BASE_REF"
fi

readarray -t target_files < <(git diff --name-only ${base_branch} --diff-filter=MA | grep "\.go")

if [[ "${#target_files[@]}" -eq 0 ]]; then
  echo "No files have changed relative to branch ${base_branch}, so there's nothing to check!"
  exit 0
fi

declare -a incorrect_files
# Array must have at least one item before we can append to it. Code below must
# work around this extra empty-string element at the beginning of the array.
incorrect_files[0]=""

for filename in "${target_files[@]}"; do
  if [[ -z "$filename" ]]; then
    continue
  fi

  output=$(go run golang.org/x/tools/cmd/goimports -l "${filename}")
  if [[ $? -ne 0 ]]; then
    echo >&2 goimports failed for "$filename"
    exit 1
  fi

  if [[ -n "$output" ]]; then
    incorrect_files+=("$filename")
  fi
done

if [[ "${#incorrect_files[@]}" -gt 1 ]]; then
  echo >&2 'The following files have import statements that disagree with "goimports"':
  for filename in "${incorrect_files[@]}"; do
    if [[ -z "$filename" ]]; then
      continue
    fi

    echo >&2 ' - ' "${filename}"
  done
  echo >&2 'Use `go run golang.org/x/tools/cmd/goimports -w -l` on each of these files to update these files.'
  exit 1
fi

echo 'All of the changed files look good!'
exit 0
