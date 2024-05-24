# Photo Deduplicator

Photo Deduplicator is a tool designed to assist Synology NAS users in deduplicating photos, a feature not inherently provided by Synology Photo. This tool leverages the PHash algorithm for accurate deduplication while ensuring safety by not auto-deleting photos, instead offering a web server for manual deletion.

## Features

- **Deduplication**: Utilizes the PHash algorithm for accurate identificatioƒn and removal of duplicate photos.
- **Safety**: Does not auto-delete photos to prevent accidental loss; provides a web server for manual deletion.
- **User-Friendly Interface**: Offers a straightforward interface for managing duplicate photos (not pretty but works).
- **Monitoring**: Monitors new photos added to the specified folder.

Note: The trash bin folder is provided to move the photo to the trash bin folder instead of deleting it outright. You can modify the launch command of the Docker image by removing --trashbin if you prefer immediate deletion. However, this is not recommended, and you do so at your own risk.

## Build from Source

To build the Photo Deduplicator from source, follow these steps:

```bash
docker build -t happyzyz/dedupphoto:latest .
```

## Usage
To use Photo Deduplicator on your Synology NAS, follow these steps:

1. Create a project using the provided `example.compose.yaml` YAML file.  
1. Start the Photo Deduplicator container.
1. Access the web server provided by Photo Deduplicator (example: `http://<nas-ip>:8989`)
1. Manually review and delete duplicate photos as needed by clicking each hash value to load and show photos.
1. (v0.2) Alternatively, you can click the "Delete All" button to automatically delete all detected duplicates, keeping only the one with the largest file size.

Note: Each image takes around 300ms to process. And you do not have to wait for all images to go through.

## License
This project is licensed under [the MIT License](https://opensource.org/license/mit). See copy of the license in this git repo.