package main

import (
	"fmt"
	htmlTemplate "html/template"
	"os"
	textTemplate "text/template"
)

type Doc struct {
	Title    string
	Path     string
	BaseName string
}

type templates struct {
	textTemplate.FuncMap
}

func (f *templates) html(filename string) (*htmlTemplate.Template, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file template (%s): %w", filename, err)
	}

	t, err := htmlTemplate.New(filename).Funcs(f.FuncMap).Parse(string(contents))
	if err != nil {
		return nil, fmt.Errorf("could not parse HTML template (%s): %w", filename, err)
	}

	return t, nil
}

func (f *templates) text(filename string) (*textTemplate.Template, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file template (%s): %w", filename, err)
	}

	t, err := textTemplate.New(filename).Funcs(f.FuncMap).Parse(string(contents))
	if err != nil {
		return nil, fmt.Errorf("could not parse text template (%s): %w", filename, err)
	}

	return t, nil
}
