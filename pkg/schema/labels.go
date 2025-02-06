package schema

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

//go:embed defaults/*
var defaults embed.FS

type StringLabel struct {
	Description string
	Key         string
	Required    bool
	Encoded     string
	Decoded     string
}

type BoolLabel struct {
	Description string
	Key         string
	Required    bool
	Decoded     bool
	Encoded     string
}

type FileLabel struct {
	Description string
	Key         string
	Required    bool
	Default     string
	Decoded     string
	Encoded     string
}

type FolderLabel struct {
	Description string
	Key         string
	Required    bool
	Files       []FileLabel
}

func (s *StringLabel) Encode(content string) error {
	if s.Required && content == "" {
		return fmt.Errorf("label %s required", s.Key)
	}

	compacted := compact([]byte(content))
	s.Decoded = string(compacted)
	s.Encoded = base64.StdEncoding.EncodeToString(compacted)
	return nil
}

func (s *StringLabel) Decode(labels map[string]string) error {
	for k, v := range labels {
		if k == s.Key {
			content, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return err
			}

			compacted := compact(content)
			s.Encoded = v
			s.Decoded = string(compacted)
			return nil
		}
	}

	if s.Required {
		return fmt.Errorf("label %s required", s.Key)
	}

	defaultPath := filepath.Join("defaults", s.Key)
	byteContent, err := defaults.ReadFile(defaultPath)
	if err != nil {
		return err
	}

	compacted := compact(byteContent)
	s.Decoded = string(compacted)
	s.Encoded = base64.StdEncoding.EncodeToString(compacted)
	return nil
}

func (s *StringLabel) Template(data ...any) (string, error) {
	return render(s.Decoded, data...)
}

func (f *FileLabel) Encode(path string) error {
	var byteContent []byte

	log.Debug().Msgf("encoding file %s", path)

	if _, err := os.Stat(path); err != nil {
		log.Debug().Msgf("file %s not found, using default", path)
		if f.Required {
			return fmt.Errorf("file %s required", path)
		}
		byteContent, err = defaults.ReadFile(f.Default)
		if err != nil {
			log.Debug().Err(err).Msgf("failed to read default file %s", path)
			return nil
		}
	} else {
		byteContent, err = os.ReadFile(path)
		if err != nil {
			return err
		}
	}

	if strings.Contains(path, ".json") {
		if !json.Valid(byteContent) {
			return fmt.Errorf("invalid JSON in %s", path)
		}
	}

	compacted := compact(byteContent)
	f.Decoded = string(compacted)
	f.Encoded = base64.StdEncoding.EncodeToString(compacted)
	return nil
}

func (f *FileLabel) Decode(labels map[string]string) error {
	for k, v := range labels {
		if k == f.Key {
			decoded, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return err
			}

			compacted := compact(decoded)
			f.Encoded = v
			f.Decoded = string(compacted)
			return nil
		}
	}

	if f.Required {
		return fmt.Errorf("label %s required", f.Key)
	}

	return nil
}

func (f *FileLabel) Template(data ...any) (string, error) {
	return render(f.Decoded, data...)
}

func (f *FolderLabel) Encode(path string) error {
	log.Debug().Msgf("encoding folder %s", path)

	if _, err := os.Stat(path); err != nil {
		if f.Required {
			return err
		}
		return nil
	}

	err := filepath.Walk(path, func(childPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// extract filename without extension(s)
			filename := filepath.Base(childPath)
			shortfilename := strings.Split(filename, ".")[0]

			// create label by combining the key and filename
			label := f.Key + "." + shortfilename

			// encode file content
			encodedFile := FileLabel{
				Description: "Individual embedded bus template",
				Key:         label,
				Required:    true,
			}

			if err := encodedFile.Encode(childPath); err != nil {
				return err
			}

			f.Files = append(f.Files, encodedFile)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if f.Required && len(f.Files) == 0 {
		return fmt.Errorf("label %s required", path)
	}

	return nil
}

func (f *FolderLabel) Decode(labels map[string]string) error {
	for k, v := range labels {
		if strings.HasPrefix(k, f.Key) {
			decodedLabel, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return err
			}

			compacted := compact(decodedLabel)
			f.Files = append(f.Files, FileLabel{
				Description: "Individual embedded bus template",
				Key:         k,
				Encoded:     v,
				Decoded:     string(compacted),
			})
		}
	}

	if f.Required && len(f.Files) == 0 {
		return fmt.Errorf("no labels with prefix %s found, but required", f.Key)
	}

	return nil
}

func (b *BoolLabel) Encode(state bool) error {
	boolString := strconv.FormatBool(state)
	encoded := base64.StdEncoding.EncodeToString([]byte(boolString))

	b.Decoded = state
	b.Encoded = encoded
	return nil
}

func (b *BoolLabel) Decode(labels map[string]string) error {
	for k, v := range labels {
		if k == b.Key {
			boolString, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return err
			}

			decoded, err := strconv.ParseBool(string(boolString))
			if err != nil {
				return err
			}

			b.Decoded = decoded
			b.Encoded = v
			return nil
		}
	}

	if b.Required {
		return fmt.Errorf("label %s required", b.Key)
	}

	return nil
}

func compact(content []byte) []byte {
	if json.Valid(content) {
		var compacted bytes.Buffer
		if err := json.Compact(&compacted, content); err != nil {
			return content
		}
		return compacted.Bytes()
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return []byte(strings.Join(lines, "\n"))
}

func render(templateContent string, data ...any) (string, error) {
	tmpl, err := template.New("document").Option("missingkey=invalid").Parse(templateContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	for _, p := range data {
		if err := tmpl.Execute(&buf, p); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}
