package yaz

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yaoapp/kun/log"
)

var defaultPatterns = []string{"*.yao", "*.json", "*.jsonc", "*.yaml", "*.so", "*.dll", "*.js", "*.py", "*.ts", "*.wasm"}

// Open opens a package file.
func Open(file string, cipher Cipher) (*Yaz, error) {

	// uncompress
	path, err := Uncompress(file)
	if err != nil {
		return nil, err
	}

	yaz := &Yaz{
		file:   file,
		cipher: cipher,
		root:   path,
	}

	return yaz, nil
}

// Walk walks the package file.
func (yaz *Yaz) Walk(root string, handler func(root, filename string, isdir bool) error, patterns ...string) error {

	rootAbs, err := yaz.abs(root)
	if err != nil {
		return err
	}

	if patterns == nil {
		patterns = defaultPatterns
	}

	return filepath.Walk(rootAbs, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error("[yaz.Walk] %s %s", filename, err.Error())
			return err
		}

		isdir := info.IsDir()
		if patterns != nil && !isdir && len(patterns) > 0 && patterns[0] != "-" {
			notmatched := true
			basname := filepath.Base(filename)
			for _, pattern := range patterns {
				if matched, _ := filepath.Match(pattern, basname); matched {
					notmatched = false
					break
				}
			}

			if notmatched {
				return nil
			}
		}

		name := strings.TrimPrefix(filename, rootAbs)
		if name == "" && isdir {
			name = string(os.PathSeparator)
		}

		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "/.") || strings.HasPrefix(name, "\\.") {
			return nil
		}

		if !isdir {
			name = filepath.Join(root, name)
		}

		err = handler(root, name, isdir)
		if err != nil {
			log.Error("[yaz.Walk] %s %s", filename, err.Error())
			return err
		}

		return nil
	})
}

// Read reads a file from the package.
func (yaz *Yaz) Read(name string) ([]byte, error) {

	file, err := yaz.abs(name)
	if err != nil {
		return nil, err
	}

	// decrypt file
	ext := strings.TrimPrefix(filepath.Ext(name), ".")
	if yaz.cipher != nil && encryptFiles[ext] {
		reader, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		buff := &bytes.Buffer{}
		err = yaz.cipher.Decrypt(reader, buff)
		if err != nil {
			return nil, err
		}
		return buff.Bytes(), nil
	}

	return os.ReadFile(file)
}

// Write writes a file to the package.
func (yaz *Yaz) Write(name string, content []byte) error {
	return fmt.Errorf("yaz is a read only filesystem")
}

// Remove removes a file from the package.
func (yaz *Yaz) Remove(name string) error {
	return fmt.Errorf("yaz is a read only filesystem")
}

// Exists checks if a file exists in the package.
func (yaz *Yaz) Exists(name string) (bool, error) {

	file, err := yaz.abs(name)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Watch watches the package file.
func (yaz *Yaz) Watch(handler func(event string, name string), interrupt chan uint8) error {
	return fmt.Errorf("yaz does not support watch")
}

func (yaz *Yaz) abs(root string) (string, error) {
	root = filepath.Join(yaz.root, root)
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return root, nil
}
