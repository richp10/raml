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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParsingSimple(t *testing.T) {
	asserter := assert.New(t)

	apiDefinition := new(APIDefinition)
	err := ParseFile("./testdata/basic.raml", apiDefinition)
	asserter.NoError(err)
	asserter.Equal("./testdata/basic.raml", apiDefinition.Filename)
	asserter.Equal("GitHub API", apiDefinition.Title)
	asserter.Equal("v3", apiDefinition.Version)
	asserter.Equal(MediaType{"application/json"}, apiDefinition.MediaType)

	asserter.Len(apiDefinition.Protocols, 1)
	asserter.Equal("HTTP", apiDefinition.Protocols[0])

	asserter.Len(apiDefinition.Documentation, 2)
	asserter.Equal("Home", apiDefinition.Documentation[0].Title)

	asserter.Equal("https://api.github.com", apiDefinition.BaseURI)
}

func TestRemoteParsing(t *testing.T) {
	asserter := assert.New(t)

	def := new(APIDefinition)
	err := ParseFile("https://raw.githubusercontent.com/demeyerthom/raml-examples/demeyerthom-patch/others/tutorial-jukebox-api/jukebox-api.raml", def)
	asserter.NoError(err)
	asserter.Equal("Jukebox API", def.Title)
	asserter.Len(def.Types, 3)
	asserter.Len(def.Types["song"].Properties, 3)
}

func TestParsingWithUriTemplate(t *testing.T) {
	asserter := assert.New(t)

	apiDefinition := new(APIDefinition)
	err := ParseFile("./testdata/uri_template.raml", apiDefinition)
	asserter.NoError(err)
	asserter.Equal("https://{subdomain}.github.com", apiDefinition.BaseURI)
	asserter.Len(apiDefinition.BaseURIParameters, 1)
	asserter.Equal("The subdomain", apiDefinition.BaseURIParameters["subdomain"].Description)
}

func TestParsingWithMediaTypeMap(t *testing.T) {
	asserter := assert.New(t)

	apiDefinition := new(APIDefinition)
	err := ParseFile("./testdata/media_types_map.raml", apiDefinition)
	asserter.NoError(err)
	asserter.Equal(MediaType{"application/json", "application/xml"}, apiDefinition.MediaType)
}

func TestParsingAnnotations(t *testing.T) {
	asserter := assert.New(t)

	apiDefinition := new(APIDefinition)
	err := ParseFile("./testdata/annotated.raml", apiDefinition)
	asserter.NoError(err)

	asserter.Len(apiDefinition.Annotations.AnnotationNames, 1)
	asserter.Len(apiDefinition.Types["Address"].Annotations.AnnotationNames, 1)
	asserter.Len(apiDefinition.SecuritySchemes["oauth_1_0"].DescribedBy.Annotations.AnnotationNames, 1)
	asserter.Len(apiDefinition.Resources["/users"].Annotations.AnnotationNames, 3)
	asserter.Len(apiDefinition.Resources["/users"].Get.Annotations.AnnotationNames, 2)
}
