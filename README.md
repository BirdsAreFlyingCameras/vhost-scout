# vhost-scout

## About

### Version Number: v0.1

vhost-scout is an adaptation of a vhost enumeration tool I wrote during an engagement, designed to be a quick, easy way to conduct vhost enumeration across multiple targets. It stores all enumerated vhosts in a SQLite database, making it easy to use the data for reports and to investigate findings.

## Installation

Download the source files from GitHub:
```bash
git clone https://github.com/BirdsAreFlyingCameras/vhost-scout
```
Enter the source files directory:
```bash
cd vhost-scout
```
Install Dependencies
```bash
go mod tidy
```
Build binary from source:
```bash
go build
```


## Usage
vhost-scout takes two positional arguments. The first argument is either a single target URL (e.g., https://example.com) or a path to a newline-separated file containing multiple targets. The second argument is a path to a newline-separated file of vhost names to test against each target.

```bash
./vhost-scout (<target.tld> | <targets.txt>) <vhosts.txt>
```
