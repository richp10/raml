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

// This file contains tests.

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TODO: Way, way more serious tests.
//
// Inspirations:
// 	 	https://github.com/raml-org/raml-js-parser/tree/master/test
//		https://github.com/raml-org/raml-java-parser/tree/master/src/test
// 		https://github.com/an2deg/pyraml-parser/tree/master/tests
//		https://github.com/cybertk/ramlev/tree/master/test/fixtures

func TestFailedParsing(t *testing.T) {

	fileNames := []string{"./samples/bad_raml.raml"}

	for _, fileName := range fileNames {

		fmt.Printf("Attempting to parse RAML file: %s\n", fileName)

		apiDef := new(APIDefinition)
		err := ParseFile(fileName, apiDef)

		if err == nil {
			t.Fatalf("Failed detecting bad RAML file %s", fileName)
		} else {
			fmt.Printf("Detected bad RAML file %s:\n%s", fileName, err.Error())
		}
	}
}

func TestParsing(t *testing.T) {

	fileNames := []string{
		"./samples/congo/api.raml",
	}

	for _, fileName := range fileNames {

		fmt.Printf("Attempting to parse RAML file: %s\n", fileName)

		apiDefinition := new(APIDefinition)
		err := ParseFile(fileName, apiDefinition)

		if err != nil {
			t.Fatalf("Failed parsing file %s:\n  %s", fileName, err.Error())
		} else {
			fmt.Printf("Successfully parsed file %s!\n", fileName)
		}
	}
}

func TestMethodStringer(t *testing.T) {
	asserter := assert.New(t)
	def := new(APIDefinition)
	err := ParseFile("./samples/simple_example.raml", def)
	asserter.NoError(err)

	r := def.Resources["/resources"]
	asserter.Equal("GET", r.Get.Name)

	n := r.Nested["/{resourceId}"]
	asserter.Equal("GET", n.Get.Name)
	asserter.Equal("PUT", n.Put.Name)
	asserter.Equal("DELETE", n.Delete.Name)
}

func TestRemoteParsing(t *testing.T) {
	asserter := assert.New(t)

	def := new(APIDefinition)
	err := ParseFile("https://raw.githubusercontent.com/raml-org/raml-tutorial-200/master/jukebox-api.raml", def)
	asserter.NoError(err)
	asserter.Equal("Jukebox API", def.Title)
	asserter.Len(def.Types, 3)
	asserter.Len(def.Types["song"].Properties, 3)
}
