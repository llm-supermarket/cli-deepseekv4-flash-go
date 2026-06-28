package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yetanotherchris/rclone-encrypt-deepseekv4-flash/internal/crypto"
	"golang.org/x/term"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type config struct {
	password        string
	salt            string
	inputFile       string
	outputFile      string
	filenameEnc     string
	encrypt         bool
	decrypt         bool
	showVersion     bool
	passwordFromEnv bool
}

func run() error {
	cfg := config{}
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--version":
			fmt.Printf("cli-deepseekv4-flash-go %s\n", version)
			return nil
		case a == "--password" || a == "-p":
			i++
			if i >= len(args) {
				return fmt.Errorf("--password requires a value")
			}
			cfg.password = args[i]
		case a == "--salt":
			i++
			if i >= len(args) {
				return fmt.Errorf("--salt requires a value")
			}
			cfg.salt = args[i]
		case a == "--input-file" || a == "-i":
			i++
			if i >= len(args) {
				return fmt.Errorf("--input-file requires a value")
			}
			cfg.inputFile = args[i]
		case a == "--output-file" || a == "-o":
			i++
			if i >= len(args) {
				return fmt.Errorf("--output-file requires a value")
			}
			cfg.outputFile = args[i]
		case a == "--filename-encoding":
			i++
			if i >= len(args) {
				return fmt.Errorf("--filename-encoding requires a value")
			}
			cfg.filenameEnc = args[i]
		case a == "encrypt":
			cfg.encrypt = true
		case a == "decrypt":
			cfg.decrypt = true
		default:
			return fmt.Errorf("unknown argument: %s", a)
		}
	}

	if !cfg.encrypt && !cfg.decrypt {
		return fmt.Errorf("must specify either 'encrypt' or 'decrypt' subcommand")
	}

	if cfg.inputFile == "" {
		return fmt.Errorf("--input-file (-i) is required")
	}

	if cfg.password == "" {
		if pw, ok := os.LookupEnv("RCLONE_ENCRYPT_PASSWORD"); ok {
			cfg.password = pw
			cfg.passwordFromEnv = true
		}
	}

	if cfg.password == "" {
		pw, err := promptPassword("Enter password: ")
		if err != nil {
			return err
		}
		cfg.password = pw

		if cfg.salt == "" {
			s, err := promptPassword("Enter salt (optional, leave empty for default): ")
			if err != nil {
				return err
			}
			cfg.salt = s
		}
	} else if !cfg.passwordFromEnv {
		fmt.Fprintln(os.Stderr, "Warning: Providing password via --password is insecure. Use RCLONE_ENCRYPT_PASSWORD environment variable instead. Consider wiping your terminal history.")
	}

	if cfg.filenameEnc == "" {
		cfg.filenameEnc = "base32"
	}

	enc, err := crypto.ParseEncoding(cfg.filenameEnc)
	if err != nil {
		return err
	}

	ciph, err := crypto.NewCipher(cfg.password, cfg.salt, enc)
	if err != nil {
		return err
	}

	if cfg.outputFile == "" {
		if cfg.encrypt {
			cfg.outputFile = cfg.inputFile + ".enc"
		} else {
			cfg.outputFile = cfg.inputFile + ".dec"
		}
	}

	inFile, err := os.Open(cfg.inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inFile.Close()

	outFile, err := os.Create(cfg.outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if cfg.decrypt {
		if err := ciph.DecryptFile(inFile, outFile); err != nil {
			return err
		}

		decName, err := ciph.DecryptFileName(stripExt(cfg.inputFile))
		if err == nil && decName != "" {
			fmt.Printf("Decrypted filename: %s\n", decName)
		}
	} else {
		if err := ciph.EncryptFile(inFile, outFile); err != nil {
			return err
		}

		encName := ciph.EncryptFileName(stripExt(cfg.outputFile))
		fmt.Printf("Encrypted filename: %s\n", encName)
	}

	fmt.Printf("Wrote output to: %s\n", cfg.outputFile)
	return nil
}

func stripExt(path string) string {
	if idx := strings.LastIndex(path, "."); idx >= 0 {
		return path[:idx]
	}
	return path
}

func promptPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	// goreleaser norestart
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(pw), nil
}

// used for testing
var stdIn io.Reader = os.Stdin
