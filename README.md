# cli-deepseekv4-flash-go
A small CLI tool that encrypts and decrypts using the rclone encryption defaults. 

Rclone uses a custom salt if no salt is provided, which this tool will use by default. A few similar tools:

- https://github.com/rclone/rclone
- https://github.com/mcolatosti/rclonedecrypt
- https://github.com/br0kenpixel/rclone-rcc
- @fyears/rclone-crypt

Rclone encryption uses: 
- NaCl SecretBox (XSalsa20 + Poly1305) for the file contents.
- AES-EME for the filenames.
- scrypt for key derivation.

## Installation

**Homebrew (macOS/Linux)**

```bash
brew tap llm-supermarket/cli-deepseekv4-flash-go https://github.com/llm-supermarket/cli-deepseekv4-flash-go
brew install cli-deepseekv4-flash-go
```

**Scoop (Windows)**

```bash
scoop bucket add cli-deepseekv4-flash-go https://github.com/llm-supermarket/cli-deepseekv4-flash-go
scoop install cli-deepseekv4-flash-go
```

## Usage

### Encrypt a file

```bash
# Interactive (password prompt)
cli-deepseekv4-flash-go encrypt -i myfile.txt -o myfile.enc

# With password flag (insecure - use env var instead)
cli-deepseekv4-flash-go encrypt -i myfile.txt -o myfile.enc --password "mypassword"

# With environment variable (recommended)
export RCLONE_ENCRYPT_PASSWORD="mypassword"
cli-deepseekv4-flash-go encrypt -i myfile.txt -o myfile.enc
```

### Decrypt a file

```bash
cli-deepseekv4-flash-go decrypt -i myfile.enc -o myfile.txt --password "mypassword"
```

### With a custom salt

```bash
cli-deepseekv4-flash-go encrypt -i myfile.txt -o myfile.enc --password "mypassword" --salt "mysalt"
```

### With filename encoding

```bash
# base32 encoding (default, rclone-compatible)
cli-deepseekv4-flash-go encrypt -i myfile.txt -o myfile.enc --filename-encoding base32

# base64 encoding
cli-deepseekv4-flash-go encrypt -i myfile.txt -o myfile.enc --filename-encoding base64
```

### Output file is optional

If `-o`/`--output-file` is omitted, the tool appends `.enc` for encryption and `.dec` for decryption:

```bash
cli-deepseekv4-flash-go encrypt -i myfile.txt
# Output: myfile.txt.enc

cli-deepseekv4-flash-go decrypt -i myfile.txt.enc
# Output: myfile.txt.enc.dec
```

## Security notes

- Using `--password` on the command line is insecure. The password may be visible in process listings and saved in your shell history.
- Prefer the `RCLONE_ENCRYPT_PASSWORD` environment variable instead.
- If you must use `--password`, consider wiping your terminal history afterwards.

## Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--password` | `-p` | Encryption password (use env var `RCLONE_ENCRYPT_PASSWORD` instead) | Prompted |
| `--salt` | | Optional salt for key derivation | rclone default salt |
| `--input-file` | `-i` | Input file path | Required |
| `--output-file` | `-o` | Output file path | `input.enc` / `input.dec` |
| `--filename-encoding` | | Filename encoding: `base32` or `base64` | `base32` |
| `--version` | | Show version | |

## Building from Source

Requires Go 1.25+.

```bash
git clone https://github.com/llm-supermarket/cli-deepseekv4-flash-go
cd cli-deepseekv4-flash-go
go build -o cli-deepseekv4-flash-go .
```

## Releases

Pushing a `vX.Y.Z` tag triggers the [Build and Release workflow](.github/workflows/build-release.yml), which cross-compiles binaries for Linux and macOS (amd64/arm64) and Windows (amd64), publishes a GitHub Release, and updates the Scoop manifest and Homebrew formula.
