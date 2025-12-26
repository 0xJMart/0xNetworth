#!/bin/bash
# Version extraction script for CI/CD
# Outputs semver-compliant version strings

set -e

# Get the git reference (tag or branch)
GIT_REF="${GITHUB_REF:-$(git rev-parse --abbrev-ref HEAD)}"
# Use short SHA (7 chars) for versions, full SHA for reference
if [ -n "$GITHUB_SHA" ]; then
  GIT_SHA_FULL="$GITHUB_SHA"
  GIT_SHA="${GITHUB_SHA:0:7}"
else
  GIT_SHA_FULL=$(git rev-parse HEAD)
  GIT_SHA=$(git rev-parse --short HEAD)
fi

# Check if this is a tag (release build)
if [[ "$GIT_REF" =~ ^refs/tags/v?([0-9]+\.[0-9]+\.[0-9]+.*)$ ]]; then
    # Extract version from tag (remove 'v' prefix if present)
    VERSION="${BASH_REMATCH[1]}"
    IS_RELEASE=true
elif [[ "$GIT_REF" =~ ^refs/tags/(.+)$ ]]; then
    # Tag but not semver format - use as-is but validate
    VERSION="${BASH_REMATCH[1]}"
    IS_RELEASE=true
else
    # Development build - use semver pre-release format
    VERSION="0.0.0-${GIT_SHA}"
    IS_RELEASE=false
fi

# Ensure version is semver compliant (basic validation)
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-.+)?$ ]]; then
    echo "Error: Version '$VERSION' is not semver compliant" >&2
    exit 1
fi

# Output version information
echo "VERSION=$VERSION"
echo "APP_VERSION=$VERSION"
echo "IS_RELEASE=$IS_RELEASE"
echo "GIT_SHA=$GIT_SHA"

# Export for use in GitHub Actions
if [ -n "$GITHUB_OUTPUT" ]; then
    echo "version=$VERSION" >> "$GITHUB_OUTPUT"
    echo "app_version=$VERSION" >> "$GITHUB_OUTPUT"
    echo "is_release=$IS_RELEASE" >> "$GITHUB_OUTPUT"
    echo "git_sha=$GIT_SHA" >> "$GITHUB_OUTPUT"
fi

