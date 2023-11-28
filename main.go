package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

func convertCurlyToColon(text string) string {
	// Replace opening curly braces with a colon and newline
	re := regexp.MustCompile(`\s*{\s*`)
	text = re.ReplaceAllString(text, ":\n")

	// Replace closing curly braces with a newline
	re = regexp.MustCompile(`\s*}\s*`)
	text = re.ReplaceAllString(text, "\n")

	// Indentation handling
	// Split the text into lines for processing
	lines := strings.Split(text, "\n")

	// Initialize variables for indentation level and the new lines list
	indentLevel := 0
	var newLines []string

	// Iterate over each line
	for _, line := range lines {
		// Check if the line is a closing brace (now an empty line after our previous replacements)
		if strings.TrimSpace(line) == "" {
			// Decrease indent level
			indentLevel--
		}
		// Add the indented line to the new lines list
		newLines = append(newLines, strings.Repeat("    ", indentLevel)+strings.TrimSpace(line))
		// Check if the line ends with a colon, indicating an opening brace
		if strings.HasSuffix(line, ":") {
			// Increase indent level
			indentLevel++
		}
	}

	// Join the new lines into a single string
	return strings.Join(newLines, "\n")
}

func duplicateAndModifyFile(originalPath, duplicatePath string) error {
	// Open the original file for reading
	content, err := ioutil.ReadFile(originalPath)
	if err != nil {
		return fmt.Errorf("error reading original file: %v", err)
	}

	// Modify the content as needed
	modifiedContent := convertCurlyToColon(string(content))

	// Get the absolute path for the duplicate file
	duplicateAbsPath, err := filepath.Abs(duplicatePath)
	if err != nil {
		return fmt.Errorf("error getting absolute path for duplicate file: %v", err)
	}

	// Open the new file for writing (the duplicated file)
	err = ioutil.WriteFile(duplicateAbsPath, []byte(modifiedContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing duplicated file: %v", err)
	}

	return nil
}

func printOutput(reader io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pythonium <python_file>")
		os.Exit(1)
	}

	originalFile := os.Args[1]
	duplicateFile := originalFile + "_run.py"

	err := duplicateAndModifyFile(originalFile, duplicateFile)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		os.Exit(1)
	}

	command := exec.Command("python", duplicateFile)

	var wg sync.WaitGroup

	// Redirect the command's output to a pipe
	stdout, err := command.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating StdoutPipe: %v\n", err)
		os.Exit(1)
	}

	// Start the command
	err = command.Start()
	if err != nil {
		fmt.Printf("Error starting command: %v\n", err)
		os.Exit(1)
	}

	// Start a goroutine to print the output
	wg.Add(1)
	go printOutput(stdout, &wg)

	// Wait for the command to finish
	err = command.Wait()
	if err != nil {
		fmt.Printf("Command '%s' failed with error: %v\n", strings.Join(command.Args, " "), err)
		os.Exit(1)
	}

	// Wait for the goroutine to finish printing
	wg.Wait()
}
