#!/bin/bash

set -e

# Define app name (base name only)
APP_NAME="rray"

# Define output base directory
OUTPUT_DIR="build"

# List of target platforms
PLATFORMS=(
  "windows/amd64"
  "windows/386"
  "windows/arm64"
  "linux/amd64"
  "linux/386"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
)

echo "🔧 Starting build..."

# Loop through platforms and build
for PLATFORM in "${PLATFORMS[@]}"
do
  GOOS="${PLATFORM%/*}"
  GOARCH="${PLATFORM#*/}"
  DEST_DIR="${OUTPUT_DIR}/${GOOS}/${GOARCH}"

  # Ensure destination directory exists
  mkdir -p "$DEST_DIR"

  # Add .exe extension for Windows
  OUTPUT_NAME="$APP_NAME"
  if [ "$GOOS" == "windows" ]; then
    OUTPUT_NAME="${APP_NAME}.exe"
  fi

  echo "📦 Building for $GOOS/$GOARCH..."

  env GOOS="$GOOS" GOARCH="$GOARCH" go build -o "${DEST_DIR}/${OUTPUT_NAME}" .

  echo "✅ Done: $DEST_DIR/$OUTPUT_NAME"
done

echo "🎉 All builds completed successfully."
