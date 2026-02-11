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

| Platform                 | Archive                                          |
|--------------------------|--------------------------------------------------|
| macOS (Apple Silicon)    | `lingua-cli-VERSION-darwin-arm64.tar.gz`         |
| macOS (Intel)            | `lingua-cli-VERSION-darwin-amd64.tar.gz`         |
| macOS (Universal Binary) | `lingua-cli-VERSION-darwin-universal.tar.gz`     |
| Linux (amd64)            | `lingua-cli-VERSION-linux-amd64.tar.gz`          |
| Linux (arm64)            | `lingua-cli-VERSION-linux-arm64.tar.gz`          |
| Windows (amd64)          | `lingua-cli-VERSION-windows-amd64.zip`           |
| Windows (arm64)          | `lingua-cli-VERSION-windows-arm64.zip`           |

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
123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890
lingua-cli is a command line tool for language classification, using the lingua-go library.

Usage: lingua-cli [OPTIONS] [TEXT]...

Arguments:
  [TEXT]... 

Options:
  -D string
        Output column delimiter. (default "\t")
  -L    List all supported languages
  -M int
        Minimum text length (without regard for whitespace, punctuation or numerals!).
        Shorter fragments will be classified as 'unknown'
  -a    Show all confidence values (entire probability distribution), rather than just
        the winning score. Does not work with --multi
  -c float
        Confidence threshold, only output results with at least this confidence value (0.0-1.0)
  -d float
        Minimum relative distance between top language probabilities (0.0-1.0).
  -l string
        Comma seperated list of iso-639-1 codes of languages to detect, if not specified,
        all supported language will be used. Setting this improves accuracy and resource usage.
  -m    Classify multiple languages in mixed texts, will return matches along with UTF-8
        byte offsets. Can not be combined with line mode.
  -n    Classify language per line, this only works if text is not supplied directly as an argument
  -q    Quick/low accuracy mode
  -version
        Print version
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
