# lingua-cli

A command-line interface for natural language detection, powered by [lingua-go](https://github.com/pemistahl/lingua-go).

This is a Go port of [lingua-cli](https://github.com/proycon/lingua-cli) (originally written in Rust using lingua-rs).

## Features

- Detects 75 languages
- Works well on short text, single words, and mixed-language text
- No external API required — runs fully offline
- Pre-built binaries available for macOS, Linux, and Windows (amd64/arm64)

## Installation

### Download a pre-built binary (recommended)

Download the appropriate archive for your platform from the Releases page:

| Platform              | Archive                                          |
|-----------------------|--------------------------------------------------|
| macOS (Apple Silicon) | `lingua-cli-VERSION-darwin-arm64.tar.gz`         |
| macOS (Intel)         | `lingua-cli-VERSION-darwin-amd64.tar.gz`         |
| Linux (amd64)         | `lingua-cli-VERSION-linux-amd64.tar.gz`          |
| Linux (arm64)         | `lingua-cli-VERSION-linux-arm64.tar.gz`          |
| Windows (amd64)       | `lingua-cli-VERSION-windows-amd64.zip`           |
| Windows (arm64)       | `lingua-cli-VERSION-windows-arm64.zip`           |

```sh
# Example: macOS Apple Silicon
curl -LO https://github.com/rinodrops/lingua-cli/releases/latest/download/lingua-cli-0.1.0-darwin-arm64.tar.gz
tar xzf lingua-cli-0.1.0-darwin-arm64.tar.gz
mv lingua-cli-0.1.0-darwin-arm64 /usr/local/bin/lingua-cli
```

### Build from source

Requires Go 1.21 or later.

```sh
git clone https://github.com/rinodrops/lingua-cli
cd lingua-cli
make build
```

## Usage

```sh
Usage: lingua-cli [OPTIONS] [TEXT...]

Options:
  -l string    Comma-separated list of ISO 639-1 language codes to detect.
               If not specified, all 75 supported languages are used.
               Restricting languages improves accuracy and reduces memory usage.
  -n           Classify language per line (stdin only).
  -L           List all supported languages and exit.
  -a           Show all confidence values (full probability distribution).
               Cannot be combined with -m.
  -q           Quick / low-accuracy mode (faster, uses less memory).
  -m           Detect multiple languages in mixed-language text.
               Returns matches with UTF-8 byte offsets.
               Cannot be combined with -n.
  -c float     Confidence threshold (0.0-1.0). Results below this value are suppressed.
  -M int       Minimum alphabetic character count. Shorter fragments are classified as 'unknown'.
  -d float     Minimum relative distance between top language probabilities (0.0-0.99).
  -D string    Output column delimiter (default: tab).
  -version     Print version and exit.
```

## Examples

**Detect language of a string:**

```sh
echo "Hello world" | lingua-cli
en      0.1437254555859748
```

**Restrict to specific languages:**

```sh
echo "Bonjour a tous" | lingua-cli -l fr,de,es,nl,en
fr      0.8115424955557187
```

**Classify line by line:**

```sh
printf "Hello\nBonjour\nHola\n" | lingua-cli -n -l en,fr,es
es      0.5291168301920521      Hello
fr      0.9416513967428743      Bonjour
es      0.5460554461673010      Hola
```

**Show all confidence values:**

```sh
echo "Hello" | lingua-cli -l en,fr,de -a
en      0.6030794987291236
fr      0.2535321997441210
de      0.1433883015267554
```

**Detect multiple languages in mixed text:**

```sh
echo "Parlez-vous francais? Ich spreche Deutsch." | lingua-cli -m -l fr,de
0       22      fr      Parlez-vous francais? 
22      43      de      Ich spreche Deutsch.
```

**Apply confidence threshold:**

```sh
echo "hi" | lingua-cli -c 0.9
unknown
```

**List supported languages:**

```sh
lingua-cli -L
```

**Text as positional argument:**

```sh
lingua-cli "This is English"
sw      0.2543307351237387
```

## Output format

### Default mode

```sh
<iso-639-1-code><delimiter><confidence>
```

### Per-line mode (-n)

```sh
<iso-639-1-code><delimiter><confidence><delimiter><original-line>
```

### Multi-language mode (-m)

```sh
<start-byte><delimiter><end-byte><delimiter><iso-639-1-code><delimiter><fragment>
```

## Building release archives

```sh
make release       # Cross-compile for all platforms and create archives in dist/
make checksums     # Generate SHA256 checksums for all archives
make clean         # Remove build artifacts
```

Archives are placed in `dist/`:

- `*.tar.gz` for macOS and Linux
- `*.zip` for Windows

## License

MIT
