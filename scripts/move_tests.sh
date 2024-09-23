#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)

source_dir="$SCRIPT_DIR/../test"
destination_dir="$SCRIPT_DIR/../../autograders"
warning="// DO NOT MODIFY THIS FILE, AS IT COMES FROM THE BUMBLEBASE REPOSITORY AND WILL BE OVERWRITTEN"

# Check if source directory exists
if [ ! -d "$source_dir" ]; then
    echo "Source directory $source_dir does not exist."
    exit 1
fi

# Check if destination directory exists
if [ ! -d "$destination_dir" ]; then
    echo "Destination directory $destination_dir does not exist."
    exit 1
fi

# Iterate over subfolders in source directory
for subdir in "$source_dir"/*; do
    if [ -d "$subdir" ]; then
        # Extract subfolder name
        subfolder_name=$(basename "$subdir")

        # Create destination directory if it doesn't exist
        destination_subdir="$destination_dir/$subfolder_name/tests"
        mkdir -p "$destination_subdir"

        # Move contents to destination directory
        mv -f "$subdir"/*.go "$destination_subdir/"

        # Prepend warning to all files in destination directory
        for file in "$destination_subdir"/*.go; do
            if [ -f "$file" ]; then
                echo "$warning" | cat - "$file" > temp && mv temp "$file"
            fi
        done

        echo "Moved contents of $subdir to $destination_subdir"
    fi
done

echo "All subfolders moved successfully."
