// Copyright 2014 DoAT. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation and/or
//    other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED “AS IS” WITHOUT ANY WARRANTIES WHATSOEVER.
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO,
// THE IMPLIED WARRANTIES OF NON INFRINGEMENT, MERCHANTABILITY AND FITNESS FOR A
// PARTICULAR PURPOSE ARE HEREBY DISCLAIMED. IN NO EVENT SHALL DoAT OR CONTRIBUTORS
// BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// // THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
// NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE,
// EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// The views and conclusions contained in the software and documentation are those of
// the authors and should not be interpreted as representing official policies,
// either expressed or implied, of DoAT.

package raml

// This file contains all of the RAML parser related code.

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

var (
	includeStringLen = len("!include ")
)

// ParseFile parses an RAML file.
// Returns a raml.APIDefinition value or an error if
// something went wrong.
func ParseFile(filePath string, root Processor) error {
	workDir, fileName := filepath.Split(filePath)
	_, err := ParseReadFile(workDir, fileName, root)
	return err
}

// ParseReadFile parse an .raml file.
// It returns API definition and the concatenated .raml file.
func ParseReadFile(workDir, fileName string, root Processor) ([]byte, error) {
	// Read original file contents into a byte array
	mainFileBytes, err := readFileOrURL(workDir, fileName)

	if err != nil {
		return []byte{}, err
	}

	// Get the contents of the main file
	mainFileBuffer := bytes.NewBuffer(mainFileBytes)

	// Verify the YAML version
	var ramlVersion string
	firstLine, err := mainFileBuffer.ReadString('\n')
	if err != nil {
		return []byte{}, fmt.Errorf("problem reading RAML file (Error: %s)", err.Error())
	}

	// We read some data...
	if len(firstLine) >= 10 {
		ramlVersion = firstLine[:10]
	}
	if ramlVersion != "#%RAML 1.0" {
		return []byte{}, errors.New("input file is not a RAML 1.0 file. Make  sure the file starts with #%RAML 1.0")
	}

	// Pre-process the original file, following !include directive
	preprocessedContentsBytes, err := preProcess(mainFileBuffer, workDir)
	if err != nil {
		return []byte{}, fmt.Errorf("error preprocessing RAML file (Error: %s)", err.Error())
	}

	// Unmarshal into an APIDefinition value
	err = yaml.Unmarshal(preprocessedContentsBytes, root)

	// Any errors?
	if err != nil {

		// Create a RAML error value
		ramlError := new(Error)

		// Copy the YAML errors into it..
		if yamlErrors, ok := err.(*yaml.TypeError); ok {
			populateRAMLError(ramlError, yamlErrors)
		} else {
			// Or just any other error, though this shouldn't happen.
			ramlError.Errors = append(ramlError.Errors, err.Error())
		}

		return []byte{}, ramlError
	}

	if err = root.PostProcess(workDir, fileName); err != nil {
		return preprocessedContentsBytes, err
	}

	// Good.
	return preprocessedContentsBytes, nil
}

// read raml file/url
func readFileOrURL(workingDir, fileName string) ([]byte, error) {
	// read from URL if it is an URL, otherwise read from local file.
	if u := strings.Join([]string{workingDir, fileName}, ""); isURL(u) {
		return readURL(u)
	}
	return readFileContents(workingDir, fileName)
}

func readURL(address string) ([]byte, error) {
	resp, err := http.Get(address)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// Reads the contents of a file, returns a bytes buffer
func readFileContents(workingDirectory string, fileName string) ([]byte, error) {

	filePath := filepath.Join(workingDirectory, fileName)

	if fileName == "" {
		return nil, fmt.Errorf("file name cannot be nil: %s", filePath)
	}

	// Read the file
	fileContentsArray, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil,
			fmt.Errorf("could not read file %s (Error: %s)", filePath, err.Error())
	}

	return fileContentsArray, nil
}

// returns true if the path is an HTTP URL
func isURL(path string) bool {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		if _, err := url.Parse(path); err == nil {
			return true
		}
	}
	return false
}

// preProcess acts as a preprocessor for a RAML document in YAML format,
// including files referenced via !include. It returns a pre-processed document.
func preProcess(originalContents io.Reader, workingDirectory string) ([]byte, error) {

	// NOTE: Since YAML doesn't support !include directives, and since go-yaml
	// does NOT play nice with !include tags, this has to be done like this.
	// I am considering modifying go-yaml to add custom handlers for specific
	// tags, to add support for !include, but for now - this method is
	// GoodEnough(TM) and since it will only happen once, I am not prematurely
	// optimizing it.

	var preprocessedContents bytes.Buffer

	// Go over each line, looking for !include tags
	scanner := bufio.NewScanner(originalContents)
	var line string

	// Scan the file until we reach EOF or error out
	for scanner.Scan() {
		line = scanner.Text()

		// Did we find an !include directive to handle?
		if idx := strings.Index(line, "!include"); idx != -1 {

			included := line[idx+includeStringLen:]

			rightOfDelimiter := strings.Join(strings.Split(included, "#")[1:], "#")
			included = strings.TrimSuffix(included, rightOfDelimiter)
			included = strings.TrimSuffix(included, "#")

			preprocessedContents.Write([]byte(line[:idx]))

			// Get the included file contents
			includedContents, err := readFileOrURL(workingDirectory, included)
			if err != nil {
				return nil, fmt.Errorf("Error including file %s:\n    %s",
					included, err.Error())
			}

			// we only parse utf8 content
			if !utf8.Valid(includedContents) {
				includedContents = []byte("")
			}

			// add newline to included content
			prepender := []byte("\n")

			// if it is in response body, we prepend "|" to make it as string
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimmedLine, "type ") || strings.HasPrefix(trimmedLine, "type:") { // in body
				prepender = []byte("|\n")
			}
			includedContents = append(prepender, includedContents...)

			// TODO: Check that you only insert .yaml, .raml, .txt and .md files
			// In case of .raml or .yaml, remove the comments
			// In case of other files, Base64 them first.

			// TODO: Better, step by step checks .. though probably it'll panic
			// Write text files in the same indentation as the first line
			internalScanner := bufio.NewScanner(bytes.NewBuffer(includedContents))

			// Indent by this much
			firstLine := true
			indentationString := ""

			// Go over each line, write it
			for internalScanner.Scan() {
				internalLine := internalScanner.Text()

				preprocessedContents.WriteString(indentationString)
				if firstLine {
					indentationString = strings.Repeat(" ", idx)
					firstLine = false
				}

				preprocessedContents.WriteString(internalLine)
				preprocessedContents.WriteByte('\n')
			}

		} else {

			// No, just a simple line.. write it
			preprocessedContents.WriteString(line)
			preprocessedContents.WriteByte('\n')
		}
	}

	// Any errors encountered?
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading YAML file: %s", err.Error())
	}
	// Return the preprocessed contents
	return preprocessedContents.Bytes(), nil
}
