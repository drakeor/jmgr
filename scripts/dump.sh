#!/bin/bash

# Output file
OUTPUT_FILE=./bin/combined.txt

# Extensions to exclude (binary/image files)
EXCLUDE_EXTENSIONS=(
    "mp3" "png" "svg" "jpg" "jpeg" "gif" "bmp" "ico"
    "mp4" "mkv" "avi" "mov" "pdf" "zip" "tar" "gz" "7z"
    "exe" "bin" "iso" "dll" "so" "o" "a" "class" "jar"
)

# Directories to exclude
EXCLUDE_DIRS=(
    ".git" "node_modules" "dist" "build" "__pycache__" "venv" ".venv" "bin"
)

# Filenames to exclude (lock files and other irrelevant large text files)
EXCLUDE_FILES=(
    "package-lock.json" 
    "npm-shrinkwrap.json" 
    "pnpm-lock.yaml" 
    "yarn.lock" 
    "Cargo.lock" 
    "composer.lock" 
    "poetry.lock" 
    ".DS_Store"
)

# Build regex patterns
EXT_PATTERN=$(IFS='|'; echo ".*\.(${EXCLUDE_EXTENSIONS[*]})\$")
FILE_PATTERN=$(IFS='|'; echo "(${EXCLUDE_FILES[*]})")

# Build -path patterns for find to exclude directories
FIND_EXCLUDE_DIRS=()
for dir in "${EXCLUDE_DIRS[@]}"; do
    FIND_EXCLUDE_DIRS+=(-path "*/$dir/*" -prune -o)
done

# Truncate the output file before starting
> "$OUTPUT_FILE"

# Find and process files
find . "${FIND_EXCLUDE_DIRS[@]}" -type f -print | while read -r FILE; do
    BASENAME=$(basename "$FILE")
    if [[ ! "$FILE" =~ $EXT_PATTERN ]] && [[ ! "$BASENAME" =~ $FILE_PATTERN ]]; then
        echo "// FILE: $FILE" >> "$OUTPUT_FILE"
        cat "$FILE" >> "$OUTPUT_FILE"
        echo -e "\n\n" >> "$OUTPUT_FILE"
    fi
done

echo "Dump completed to $OUTPUT_FILE"