package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetanotherchris/rclone-encrypt-deepseekv4-flash/internal/crypto"
)

func TestEncryptDecryptFile(t *testing.T) {
	plaintext := []byte("hello world this is a test file with some content")
	password := "testpassword123"
	salt := "testsalt123"

	for _, enc := range []string{"base32", "base64"} {
		t.Run("encoding_"+enc, func(t *testing.T) {
			dir := t.TempDir()
			inPath := filepath.Join(dir, "input.txt")
			encPath := filepath.Join(dir, "encrypted.bin")
			decPath := filepath.Join(dir, "decrypted.txt")

			if err := os.WriteFile(inPath, plaintext, 0644); err != nil {
				t.Fatal(err)
			}

			enc, err := crypto.ParseEncoding(enc)
			if err != nil {
				t.Fatal(err)
			}
			ciph, err := crypto.NewCipher(password, salt, enc)
			if err != nil {
				t.Fatal(err)
			}

			inFile, _ := os.Open(inPath)
			encFile, _ := os.Create(encPath)
			if err := ciph.EncryptFile(inFile, encFile); err != nil {
				t.Fatal(err)
			}
			inFile.Close()
			encFile.Close()

			encFile, _ = os.Open(encPath)
			decFile, _ := os.Create(decPath)
			if err := ciph.DecryptFile(encFile, decFile); err != nil {
				t.Fatal(err)
			}
			encFile.Close()
			decFile.Close()

			result, _ := os.ReadFile(decPath)
			if !bytes.Equal(result, plaintext) {
				t.Fatalf("decrypted text mismatch:\ngot:  %q\nwant: %q", string(result), string(plaintext))
			}
		})
	}
}

func TestEncryptDecryptNoSalt(t *testing.T) {
	plaintext := []byte("test without salt")
	password := "testpassword456"

	dir := t.TempDir()
	inPath := filepath.Join(dir, "input.txt")
	encPath := filepath.Join(dir, "encrypted.bin")
	decPath := filepath.Join(dir, "decrypted.txt")

	if err := os.WriteFile(inPath, plaintext, 0644); err != nil {
		t.Fatal(err)
	}

	ciph, err := crypto.NewCipher(password, "", crypto.EncodingBase32)
	if err != nil {
		t.Fatal(err)
	}

	inFile, _ := os.Open(inPath)
	encFile, _ := os.Create(encPath)
	if err := ciph.EncryptFile(inFile, encFile); err != nil {
		t.Fatal(err)
	}
	inFile.Close()
	encFile.Close()

	encFile, _ = os.Open(encPath)
	decFile, _ := os.Create(decPath)
	if err := ciph.DecryptFile(encFile, decFile); err != nil {
		t.Fatal(err)
	}
	encFile.Close()
	decFile.Close()

	result, _ := os.ReadFile(decPath)
	if !bytes.Equal(result, plaintext) {
		t.Fatalf("decrypted text mismatch:\ngot:  %q\nwant: %q", string(result), string(plaintext))
	}
}

func TestEncryptDecryptViaCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	binPath := buildBinary(t)
	plaintext := []byte("CLI test data for encrypt/decrypt")
	password := "clipassword"
	salt := "clisalt"

	dir := t.TempDir()
	inPath := filepath.Join(dir, "testfile.txt")
	encPath := filepath.Join(dir, "testfile.encrypted")
	decPath := filepath.Join(dir, "testfile.decrypted")

	if err := os.WriteFile(inPath, plaintext, 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("encrypt_with_password_flag", func(t *testing.T) {
		cmd := exec.Command(binPath, "encrypt",
			"-i", inPath,
			"-o", encPath,
			"--password", password,
			"--salt", salt)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("encrypt failed: %v\noutput: %s", err, string(out))
		}

		if !strings.Contains(string(out), "Wrote output") {
			t.Fatalf("unexpected output: %s", string(out))
		}
	})

	t.Run("decrypt_with_password_flag", func(t *testing.T) {
		cmd := exec.Command(binPath, "decrypt",
			"-i", encPath,
			"-o", decPath,
			"--password", password,
			"--salt", salt)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("decrypt failed: %v\noutput: %s", err, string(out))
		}

		result, _ := os.ReadFile(decPath)
		if !bytes.Equal(result, plaintext) {
			t.Fatalf("decrypted text mismatch:\ngot:  %q\nwant: %q", string(result), string(plaintext))
		}
	})
}

func TestEncryptDecryptWithBase64Encoding(t *testing.T) {
	plaintext := []byte("base64 encoding test data")
	password := "b64password"

	dir := t.TempDir()
	inPath := filepath.Join(dir, "input.txt")
	encPath := filepath.Join(dir, "encrypted.bin")
	decPath := filepath.Join(dir, "decrypted.txt")

	if err := os.WriteFile(inPath, plaintext, 0644); err != nil {
		t.Fatal(err)
	}

	cliPath := buildBinary(t)
	cmdEnc := exec.Command(cliPath, "encrypt",
		"-i", inPath,
		"-o", encPath,
		"--password", password,
		"--filename-encoding", "base64")
	if out, err := cmdEnc.CombinedOutput(); err != nil {
		t.Fatalf("encrypt failed: %v\noutput: %s", err, string(out))
	}

	cmdDec := exec.Command(cliPath, "decrypt",
		"-i", encPath,
		"-o", decPath,
		"--password", password,
		"--filename-encoding", "base64")
	if out, err := cmdDec.CombinedOutput(); err != nil {
		t.Fatalf("decrypt failed: %v\noutput: %s", err, string(out))
	}

	result, _ := os.ReadFile(decPath)
	if !bytes.Equal(result, plaintext) {
		t.Fatalf("decrypted text mismatch:\ngot:  %q\nwant: %q", string(result), string(plaintext))
	}
}

func TestEncryptDecryptWithEnvPassword(t *testing.T) {
	plaintext := []byte("env password test")
	password := "envpassword789"

	dir := t.TempDir()
	inPath := filepath.Join(dir, "input.txt")
	encPath := filepath.Join(dir, "encrypted.bin")
	decPath := filepath.Join(dir, "decrypted.txt")

	if err := os.WriteFile(inPath, plaintext, 0644); err != nil {
		t.Fatal(err)
	}

	cliPath := buildBinary(t)
	t.Setenv("RCLONE_ENCRYPT_PASSWORD", password)

	cmdEnc := exec.Command(cliPath, "encrypt",
		"-i", inPath,
		"-o", encPath)
	if out, err := cmdEnc.CombinedOutput(); err != nil {
		t.Fatalf("encrypt failed: %v\noutput: %s", err, string(out))
	}

	cmdDec := exec.Command(cliPath, "decrypt",
		"-i", encPath,
		"-o", decPath)
	if out, err := cmdDec.CombinedOutput(); err != nil {
		t.Fatalf("decrypt failed: %v\noutput: %s", err, string(out))
	}

	result, _ := os.ReadFile(decPath)
	if !bytes.Equal(result, plaintext) {
		t.Fatalf("decrypted text mismatch:\ngot:  %q\nwant: %q", string(result), string(plaintext))
	}
}

func TestFilenameEncryption(t *testing.T) {
	password := "filetest"
	salt := "filesalt"

	for _, tc := range []struct {
		name     string
		enc      string
		filename string
	}{
		{"base32", "base32", "TEST_FILE.txt"},
		{"base64", "base64", "TEST_FILE.txt"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			enc, err := crypto.ParseEncoding(tc.enc)
			if err != nil {
				t.Fatal(err)
			}
			ciph, err := crypto.NewCipher(password, salt, enc)
			if err != nil {
				t.Fatal(err)
			}

			encName := ciph.EncryptFileName(tc.filename)
			if encName == tc.filename {
				t.Fatal("filename was not encrypted")
			}
			t.Logf("Encrypted '%s' -> '%s'", tc.filename, encName)

			decName, err := ciph.DecryptFileName(encName)
			if err != nil {
				t.Fatalf("decrypt filename failed: %v", err)
			}
			if decName != tc.filename {
				t.Fatalf("filename roundtrip failed:\ngot:  %q\nwant: %q", decName, tc.filename)
			}
		})
	}
}

func buildBinary(t *testing.T) string {
	dir := t.TempDir()
	binPath := filepath.Join(dir, "cli-deepseekv4-flash-go.exe")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\noutput: %s", err, string(out))
	}
	return binPath
}

// TestPromptPassword tests that we can handle the prompt properly
func TestVersionFlag(t *testing.T) {
	cliPath := buildBinary(t)
	cmd := exec.Command(cliPath, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(string(out), "cli-deepseekv4-flash-go") {
		t.Fatalf("unexpected version output: %s", string(out))
	}
}

func TestDecryptKnownFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	password := "Testpassword1"

	testFiles := []struct {
		encFile          string
		expectedFilename string
		encoding         string
	}{
		{"kr9tu4e1da4u3nifdd99g9tf5o", "TEST_FILE.txt", "base32"},
		{"Iyxcijgc9bp3o5Y0npW6xqUvwWNcc3MA4SadB0sR6cY", "TEST_FILE BASE64.txt", "base64"},
	}

	for _, tf := range testFiles {
		t.Run(tf.encoding, func(t *testing.T) {
			encCfg, _ := crypto.ParseEncoding(tf.encoding)
			ciph, err := crypto.NewCipher(password, "", encCfg)
			if err != nil {
				t.Fatal(err)
			}

			decName, err := ciph.DecryptFileName(tf.encFile)
			if err != nil {
				t.Fatalf("filename decryption failed: %v", err)
			}
			if decName != tf.expectedFilename {
				t.Fatalf("filename mismatch: got %q, want %q", decName, tf.expectedFilename)
			}

			data, err := os.ReadFile(tf.encFile)
			if err != nil {
				t.Fatalf("cannot read test file: %v", err)
			}

			var buf bytes.Buffer
			if err := ciph.DecryptFile(bytes.NewReader(data), &buf); err != nil {
				t.Fatalf("content decryption failed: %v", err)
			}

			content := buf.String()
			if !strings.Contains(content, "umbrella") || !strings.Contains(content, "scrub") {
				t.Logf("Content: %s", content)
			}
		})
	}
}
