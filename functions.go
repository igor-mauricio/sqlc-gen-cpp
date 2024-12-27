package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

func getFuncMap(schemaParam []string) template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"oneLineSQL": func(s string) string {
			// Remove inline comments (e.g., "-- comment")
			re := regexp.MustCompile(`--.*`)
			withoutComments := re.ReplaceAllString(s, "")
			// Replace newlines and extra spaces with a single space
			cleaned := strings.ReplaceAll(withoutComments, "\n", " ")
			cleaned = strings.Join(strings.Fields(cleaned), " ") // Removes extra spaces

			return cleaned
		},

		"schema": func() string {
			schema := ""
			for _, schemaPath := range schemaParam {
				fileInfo, err := os.Stat(schemaPath)
				if err != nil {
					log.Fatalf("Error accessing schema path: %v", err)
				}
				if !fileInfo.IsDir() {
					content, err := readFile(schemaPath)
					if err != nil {
						log.Fatalf("Error reading file %s: %v", schemaPath, err)
					}
					schema += content + "\n"
					continue
				}
				files, err := os.ReadDir(schemaPath)
				if err != nil {
					log.Fatalf("Error reading directory: %v", err)
				}
				for _, file := range files {
					if file.IsDir() { // Skip subdirectories
						continue
					}
					filePath := filepath.Join(schemaPath, file.Name())
					content, err := readFile(filePath)
					if err != nil {
						log.Fatalf("Error reading file %s: %v", filePath, err)
					}
					schema += content + "\n"
				}
			}
			return schema
		},
		"regexMatch": func(pattern, s string) bool {
			match, err := regexp.MatchString(pattern, s)
			if err != nil {
				log.Fatalf("Error matching regex: %v", err)
			}
			return match
		},
	}
}

func readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
