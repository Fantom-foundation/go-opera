#!/usr/bin/env bash
set -e # Exit with nonzero exit code if anything fails

SOURCE_BRANCH="master"
TARGET_BRANCH="gh-pages"

REPO=$(git config remote.origin.url)
SHA=$(git rev-parse --verify HEAD)
HTTPS_REPO=${REPO/https:\/\/github.com\//https://${GITHUB_USER}:${GITHUB_TOKEN}@github.com/}
OUT_DIR="cloned-gh-pages"

echo "Repo: " ${REPO}
echo "HTTPS_REPO: " ${HTTPS_REPO}

# Pull requests and commits to other branches shouldn't try to deploy, just build to verify
if [[ ${TRAVIS_PULL_REQUEST} != "false" || ${TRAVIS_BRANCH} != ${SOURCE_BRANCH} ]]; then
    echo "Skipping deploy; just doing a build."
    exit 0
fi

# Clone the existing gh-pages for this repo into gh-pages/
# Create a new empty branch if gh-pages doesn't exist yet (should only happen on first deploy)
git clone ${HTTPS_REPO} ${OUT_DIR}

echo "Entering gh-pages output folder '${OUT_DIR}'"
cd ${OUT_DIR}
git checkout ${TARGET_BRANCH} || git checkout --orphan ${TARGET_BRANCH}

# Clean out existing contents
git rm -rf . || exit 0

echo "currently in dir: " $(pwd)
ls -l
git status

SPHINX_BUILD_HTML=../_build/html
echo "Copying Sphinx html from "
ls -l ${SPHINX_BUILD_HTML}
cp -R ${SPHINX_BUILD_HTML}/* .

echo "gh-pages contents now looks like:"
ls -la

git config user.name "Travis CI"
git config user.email "travis@travis-ci.org"

echo "adding .nojekyll file"
touch .nojekyll

git add .
git status

# Commit and push changes using $GITHUB_TOKEN
git commit -m "Deploy to Github Pages from commit: ${SHA}"
git status
echo "Pushing changes to ${HTTPS_REPO} ${TARGET_BRANCH} from dir $(pwd)"
git push --set-upstream origin ${TARGET_BRANCH}

echo "Done updating gh-pages"
