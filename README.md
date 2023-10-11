# BOOM Package Manager

BOOM is a simple package manager for managing software on your system. It provides commands to handle the installation, execution and maintenance of programs.

## Installation

To use BOOM you'll need to install it on your system. Here's how:

1. **Download the Binary:**
   - Visit the [BOOM Releases](https://github.com/jooapa/boom/releases) page on GitHub.
   - Download the appropriate binary for your operating system and architecture.

2. **Install BOOM:**
   - Place the downloaded binary in a directory included in your system's PATH. This allows you to run `boom` from any location in your terminal.

3. **Initialization:**
   - Before you can start using BOOM, you need to initialize it by running:
     ```bash
     boom init
     ```

## Usage

BOOM provides various commands to manage your programs. Here's a quick overview of the available commands:

```bash
Usage: boom <command> [arguments]

Commands:
  version   Display the BOOM version
  run       Run a program
  install   Install a program
  uninstall Uninstall a program
  update    Update a program
  list      List all installed programs
  search    Search for a program
  init      Initialize BOOM
  start     open BOOM in File Explorer

```
## Installation Directory

BOOM installs programs in the USER directory under a hidden .boom folder *( if using linux )*. Here's a breakdown of the directory structure:

**~/.boom/** - This is the main BOOM directory located in the user's home directory.

**installed.json** - This JSON file keeps track of all the programs installed using BOOM. It contains information about the installed programs, such as their names, versions, and installation paths.

**programs/** - This directory stores the actual software programs that you install using BOOM. Each program has its own subdirectory here.