# Change Log
All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](http://semver.org/).

Releases before this file was introduced are documented in the [GitHub releases](https://github.com/passbolt/go-passbolt-cli/releases).

## [Unreleased]
### Fixed
- PB-53937: Sign shared v5 metadata with both the user key and the metadata key (SDK bump to v0.8.3)

### Maintenance
- PB-53937: Sign the metadata private key payload with a timestamp when bootstrapping the test environment
