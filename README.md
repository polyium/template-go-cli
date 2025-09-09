<!-- {{ $.source.start }} -->

## Overview

`go` excels in its standard-library, cross system compilation, and especially relating to dependencies, vendoring and binary sizes.

The following project is a template for commandline applications. Additionally, the template options for Homebrew.

<details>
<summary>Template</summary>

<!-- {{ $.source.end }} -->

<!-- {{ $.content.start }} -->

# {{ $.content["title"] }}

> [!CAUTION]
> The following repository is in very early stages of development and is not ready for use. *The **Caution** banner will be removed when generally available*.

## Getting Started

### Installing

```bash
brew tap {{ default $.content["homebrew-taps"] "polyium/homebrew-taps" }}

brew install {{ default $.content["executable"] "template-go-cli" }}
```

###### Upgrading

```bash
brew update && brew upgrade {{ default $.content["executable"] "template-go-cli" }}
```

###### Reinstalling

```bash
brew update
brew uninstall --cask {{ default $.content["executable"] "template-go-cli" }}
brew install {{ default $.content["executable"] "template-go-cli" }}
```

## Usage

```bash
template-go-cli --help
```

## Contributions

See the [**Contributing Guide**](./CONTRIBUTING.md) for additional details on getting started.

<!-- {{ $.content.end }} -->

</details>
