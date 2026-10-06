#!/usr/bin/env bash
set -e

# SiteCraft Starter Kit Project Initializer
# Usage: ./init-project.sh [--name="my-site"] [--preset=medical] [--theme=medical]

NAME="sitecraft-app"
PRESET="legal"
THEME="legal"

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --name=*) NAME="${1#*=}" ;;
        --preset=*) PRESET="${1#*=}" ;;
        --theme=*) THEME="${1#*=}" ;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

echo "🚀 Initializing SiteCraft Project: $NAME"
echo "📦 Preset: $PRESET"
echo "🎨 Theme:  $THEME"

# 1. Compile templates
echo "==> Compiling Templ templates..."
templ generate

# 2. Compile Tailwind CSS
echo "==> Building Tailwind CSS styles..."
tailwindcss -i ui/assets/input.css -o ui/assets/styles.css --minify

# 3. Build Go Binary
echo "==> Compiling Single Go Binary..."
go build -ldflags="-s -w" -o "bin/$NAME" .

echo "✅ Project successfully built: bin/$NAME"
echo ""
echo "To launch your server with preset applied:"
echo "  ./bin/$NAME serve --http=127.0.0.1:8090 --preset=$PRESET"
echo ""
echo "PocketBase Admin UI will be available at: http://127.0.0.1:8090/_/"
