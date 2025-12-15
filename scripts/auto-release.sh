#!/bin/bash
# Auto-release script for semantic versioning

set -e

# Check if bump type is provided
if [ -z "$1" ]; then
  echo "Usage: $0 [major|minor|patch]"
  exit 1
fi

BUMP_TYPE=$1

# Read current version
if [ ! -f VERSION ]; then
  echo "0.0.0" > VERSION
fi

CURRENT_VERSION=$(cat VERSION)
echo "Current version: $CURRENT_VERSION"

# Parse version components
IFS='.' read -r -a VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR=${VERSION_PARTS[0]}
MINOR=${VERSION_PARTS[1]}
PATCH=${VERSION_PARTS[2]}

# Bump version
case $BUMP_TYPE in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch)
    PATCH=$((PATCH + 1))
    ;;
  *)
    echo "Invalid bump type: $BUMP_TYPE"
    echo "Use: major, minor, or patch"
    exit 1
    ;;
esac

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
echo "New version: $NEW_VERSION"

# Update VERSION file
echo "$NEW_VERSION" > VERSION

# Update CHANGELOG.md
DATE=$(date +%Y-%m-%d)
TEMP_CHANGELOG=$(mktemp)

# Create new changelog entry
{
  echo "# Changelog"
  echo ""
  echo "All notable changes to this project will be documented in this file."
  echo ""
  echo "The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),"
  echo "and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)."
  echo ""
  echo "## [$NEW_VERSION] - $DATE"
  echo ""

  # Add section based on bump type
  case $BUMP_TYPE in
    major)
      echo "### Breaking Changes"
      echo ""
      echo "- Major version update"
      echo ""
      ;;
    minor)
      echo "### Added"
      echo ""
      echo "- New features and improvements"
      echo ""
      ;;
    patch)
      echo "### Fixed"
      echo ""
      echo "- Bug fixes and improvements"
      echo ""
      ;;
  esac

  # Append old changelog (skip header lines)
  if [ -f CHANGELOG.md ]; then
    tail -n +8 CHANGELOG.md
  fi
} > "$TEMP_CHANGELOG"

mv "$TEMP_CHANGELOG" CHANGELOG.md

echo "Updated CHANGELOG.md"

# Create git commit and tag
git add VERSION CHANGELOG.md
git commit -m "chore: Bump version to $NEW_VERSION"
git tag -a "v$NEW_VERSION" -m "Release v$NEW_VERSION"

echo "Created commit and tag for v$NEW_VERSION"
echo "Run 'git push origin main && git push origin v$NEW_VERSION' to publish"
