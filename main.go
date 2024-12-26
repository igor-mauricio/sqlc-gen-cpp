package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/sqlc-dev/plugin-sdk-go/codegen"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func main() {
	codegen.Run(generate)
}

type Options struct {
	Template         string `json:"template" yaml:"template"`
	Filename         string `json:"filename" yaml:"filename"`
	FormatterCommand string `json:"formatter_cmd" yaml:"formatter_cmd"`
	Out              string `json:"out" yaml:"out"`
}

func parseOpts(req *plugin.GenerateRequest) (*Options, error) {
	var options Options
	if len(req.PluginOptions) == 0 {
		return &options, nil
	}
	if err := json.Unmarshal(req.PluginOptions, &options); err != nil {
		return nil, fmt.Errorf("unmarshalling plugin options: %w", err)
	}

	return &options, nil
}

func generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	options, _ := parseOpts(req)

	templateFileName := options.Template

	funcMap := template.FuncMap{
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
			for _, schemaPath := range req.Settings.Schema {
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
	}

	tmpl, err := template.New(templateFileName).Funcs(funcMap).ParseFiles(templateFileName)
	if err != nil {
		log.Fatalf("Error parsing template file: %v", err)
	}

	resp := plugin.GenerateResponse{}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, req)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	if options.FormatterCommand != "" {
		execCommand := exec.Command("/usr/bin/env", "bash", "-c", options.FormatterCommand)
		execCommand.Stdin = bytes.NewReader(buf.Bytes())
		var output bytes.Buffer
		execCommand.Stdout = &output
		if err := execCommand.Run(); err != nil {
			log.Fatalf("Error executing formatter command: %v", err)
		}

		buf = output
	}

	resp.Files = append(resp.Files, &plugin.File{
		Name:     options.Filename,
		Contents: buf.Bytes(),
	})

	return &resp, nil
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
