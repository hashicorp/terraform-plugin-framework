package main

import (
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

func main() {
	path, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		os.Exit(1)
	}
	fmt.Println("Current directory:", path)

	features := struct {
		Optionality       bool
		Computedness      bool
		Sensitivity       bool
		WriteOnly         bool
		RequiredForImport bool
		Defaultiness      bool
		Planning          bool
	}{
		true,  // Optionality
		false, // Computedness
		false, // Sensitivity
		false, // WriteOnly
		false, // RequiredForImport
		false, // Defaultiness
		false, // Planning
	}

	t, err := template.ParseGlob("generate/*.go.tmpl")
	if err != nil {
		fmt.Println("Error parsing templates:", err)
		os.Exit(1)
	}

	file, err := os.OpenFile(path+"/string_attribute.go", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	err = t.ExecuteTemplate(file, "string_attribute.go.tmpl", features)
	if err != nil {
		fmt.Println("Error executing template:", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "fmt", path+"/string_attribute.go")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error formatting file:", err)
		os.Exit(1)
	}
	fmt.Println("File generated and formatted successfully: string_attribute.go")
}
