# Photo Deduplicator

Photo Deduplicator is a tool designed to assist Synology NAS users in deduplicating photos, a feature not inherently provided by Synology Photo. This tool leverages the PHash algorithm for accurate deduplication while ensuring safety by not auto-deleting photos, instead offering a web server for manual deletion.

## Features

- **Deduplication**: Utilizes the PHash algorithm for accurate identificatioƒn and removal of duplicate photos.
- **Safety**: Does not auto-delete photos to prevent accidental loss; provides a web server for manual deletion.
- **User-Friendly Interface**: Offers a straightforward interface for managing duplicate photos (not pretty but works).

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
1. Manually review and delete duplicate photos as needed.

## License
This project is licensed under [the MIT License](https://opensource.org/license/mit). See copy of the license in this git repo.