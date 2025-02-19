# s3-cp-tagging

[![Test](https://github.com/urth-inc/s3-cp-tagging/actions/workflows/test.yml/badge.svg)](https://github.com/urth-inc/s3-cp-tagging/actions/workflows/test.yml)

A command-line tool to copy local files to Amazon S3 with tagging support. Similar to the AWS CLI's `s3 cp` command but with the ability to add tags during upload.

## Why this tool?

The official AWS CLI's `s3 cp` command does not support setting tags during the upload operation. While AWS S3 supports object tagging, you would need to make a separate API call to add tags after uploading files. This tool combines these operations into a single command, making it more efficient to upload files with tags in one step.

## Features

- Upload files and directories to S3 with tags
- AWS profile support
- Recursive directory upload
- File filtering with exclude/include patterns (like AWS CLI)
- Similar interface to AWS CLI's `s3 cp` command

## Usage

```bash
# Upload a directory with tags
s3-cp-tagging ./dist s3://my-bucket --tagging "TagSet=[{Key=version,Value=v1.0.0},{Key=commit,Value=abc123}]"

# Use a specific AWS profile
s3-cp-tagging ./dist s3://my-bucket --profile myprofile --tagging "TagSet=[{Key=version,Value=v1.0.0}]"

# Exclude .map files
s3-cp-tagging ./dist s3://my-bucket --exclude "*.map" --tagging "TagSet=[{Key=version,Value=v1.0.0}]"

# Exclude all but JavaScript files
s3-cp-tagging ./dist s3://my-bucket --exclude "*" --include "*.js" --tagging "TagSet=[{Key=version,Value=v1.0.0}]"
```

### Options

- `--profile`: Use a specific profile from your AWS credentials file
- `--tagging`: The tag-set for the object. Must be formatted as: `"TagSet=[{Key=key1,Value=value1},{Key=key2,Value=value2}]"`
- `--exclude`: Exclude files that match the specified pattern (can be specified multiple times)
- `--include`: Don't exclude files that match the specified pattern (can be specified multiple times)

## Prerequisites

- Go 1.24 or later
- AWS credentials configured
- Make (for building)

## Installation

### From Source

1. Clone the repository:
```bash
git clone https://github.com/urth-inc/s3-cp-tagging.git
cd s3-cp-tagging
```

2. Build the binary:
```bash
make build
```

The binary will be available in the `bin` directory.

## Development

### Setup Development Environment

1. Install dependencies:
```bash
go mod download
```

2. Run tests:
```bash
make test
```

3. Run tests with coverage:
```bash
make coverage
```

### Build

- Build for development:
```bash
make build
```

- Build with version information:
```bash
make release
```

## License

MIT License
