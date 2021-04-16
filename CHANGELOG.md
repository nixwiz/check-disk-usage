# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

## [0.6.0] - 2021-04-16

### Added
- Added --inode-mode option to monitore inode information instead of block usage (df -i)

## [0.4.2] - 2021-03-31

### Changed
- Fixed typo(s)
- README update(s)

## [0.4.1] - 2021-02-14

### Changed
- Fixed README re: read-only

## [0.4.0] - 2021-02-14

### Added
- Added --include-read-only option to support ignoring read-only file systems

## [0.3.1] - 2020-12-30

### Changed
- Added newline to end of UNKNOWN error output

## [0.3.0] - 2020-12-30

### Changed
- Moved included/excluded checks against partition information up before querying usage

## [0.2.0] - 2020-12-30

### Added
- Added flags --include-pseudo-fs and --fail-on-error

### Changed
- Changed the behavior for failures to get fs usage based on flag
- Changed call to get partitions to use --include-pseudo-fs flag

## [0.1.1] - 2020-12-29

### Change
- Added appropriate error checking

## [0.1.0] - 2020-12-14

### Added
- Initial release
