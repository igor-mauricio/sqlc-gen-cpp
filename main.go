package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	options, err := parseOpts(req)
	if err != nil {
		return nil, fmt.Errorf("parsing plugin options: %w", err)
	}
	templatePathName := options.Template

	resp := plugin.GenerateResponse{}

	//check if the path is a directory

	fileInfo, err := os.Stat(templatePathName)
	if err != nil {
		log.Fatalf("Error accessing template path: %v", err)
	}

	if !fileInfo.IsDir() {
		tmpl, err := template.New(templatePathName).Funcs(getFuncMap(req.Settings.Schema)).ParseFiles(templatePathName)
		if err != nil {
			log.Fatalf("Error parsing template file: %v", err)
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, req)
		if err != nil {
			log.Fatalf("Error executing template: %v", err)
		}
		// runFormatter(&buf, options.FormatterCommand)
		resp.Files = append(resp.Files, &plugin.File{
			Name:     strings.TrimSuffix(templatePathName, ".tmpl"),
			Contents: buf.Bytes(),
		})
		return &resp, nil
	}
	files, err := os.ReadDir(templatePathName)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}
	for _, file := range files {
		relativePath := fmt.Sprintf("%s/%s", templatePathName, file.Name())
		fileName := file.Name()
		if file.IsDir() {
			continue
		}
		tmpl, err := template.New(fileName).Funcs(getFuncMap(req.Settings.Schema)).ParseFiles(relativePath)
		if err != nil {
			log.Fatalf("Error parsing template file: %v", err)
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, req)
		if err != nil {
			log.Fatalf("Error executing template: %v", err)
		}
		// runFormatter(&buf, options.FormatterCommand)
		resp.Files = append(resp.Files, &plugin.File{
			Name:     strings.TrimSuffix(fileName, ".tmpl"),
			Contents: buf.Bytes(),
		})
	}

	return &resp, nil
}

func runFormatter(buffer *bytes.Buffer, formatterCommand string) {
	execCommand := exec.Command("/usr/bin/env", "bash", "-c", formatterCommand)
	execCommand.Stdin = bytes.NewReader(buffer.Bytes())
	var output bytes.Buffer
	execCommand.Stdout = &output
	if err := execCommand.Run(); err != nil {
		log.Fatalf("Error executing formatter command: %v", err)
	}

	*buffer = output
}
