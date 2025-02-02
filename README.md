# GLS - Go-based LS Package

## Overview

gls is ls but in Go. It provides a variety of options to customize the display of files and directories, including sorting, filtering, and previewing file contents, and extra features traditional ls does not provide.

## Features

- **Sorting**: Sort files by name, size, or date.
- **Filtering**: Show only directories, files, or hidden files.
- **Preview**: Preview the contents of files directly in the terminal.
- **Interactive Mode**: Navigate and manage files using a TUI (Text User Interface).
- **File Operations**: Rename files, create archives, and delete files.

## Installation

To install GLS, use the following command:

```sh
go get github.com/rinimisini112/gls
```

Build and move to your path

```sh
go build -o gls github.com/rinimisini112/gls
sudo mv gls /usr/local/bin
```

## Usage

```sh
gls [options] [directories]
```

### Options

- `-s=name`      Sort by name (default)
- `-s=size`      Sort by size
- `-s=date`      Sort by date
- `-a`           Show hidden files
- `-p`           Preview files
- `-t=dir`       Show only directories
- `-t=file`      Show only files
- `-t=hidden`    Show only hidden files
- `-l [limit]`   Limit the number of files displayed
- `-s [query]`   Search for files containing 'query'
- `--help, -h`   Show help message
- `--rename [old] [new]` Rename a file

## Example

```sh
gls -s=size -a -p /home/user/documents
```

This command will list all files in the `/home/user/documents` directory, sorted by size, including hidden files, and previewing the contents of each file.

## Interactive Mode

To start the interactive mode, use the `-i` or `--interactive` option:

```sh
gls -i /home/user/documents
```

In interactive mode, you can navigate through files and directories, preview file contents, and perform file operations using keyboard shortcuts.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the MIT License.