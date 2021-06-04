package raml

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSingularize(t *testing.T) {
	asserter := assert.New(t)
	var tests = []struct {
		Word   string
		Result string
	}{
		{"Users", "User"},
		{"Books", "Book"},
		{"Potatoes", "Potato"},
		{"Students", "Student"},
	}

	for _, test := range tests {
		asserter.Equal(test.Result, singularize(test.Word))
	}
}

func TestPluralize(t *testing.T) {
	asserter := assert.New(t)
	var tests = []struct {
		Word   string
		Result string
	}{
		{"User", "Users"},
		{"Book", "Books"},
		{"Potato", "Potatoes"},
		{"Student", "Students"},
	}

	for _, test := range tests {
		asserter.Equal(test.Result, pluralize(test.Word))
	}
}

func TestLowerCamelCase(t *testing.T) {
	asserter := assert.New(t)

	var tests = []struct {
		Word   string
		Result string
	}{
		{"userId", "userId"},
		{"UserId", "userId"},
		{"user_id", "userId"},
		{"user-id", "userId"},
	}

	for _, test := range tests {
		asserter.Equal(lowerCamelCase(test.Word), test.Result)
	}
}

func TestUpperCamelCase(t *testing.T) {
	asserter := assert.New(t)

	var tests = []struct {
		Word   string
		Result string
	}{
		{"userId", "UserId"},
		{"UserId", "UserId"},
		{"user_id", "UserId"},
		{"user-id", "UserId"},
	}

	for _, test := range tests {
		asserter.Equal(upperCamelCase(test.Word), test.Result)
	}
}

func TestLowerUnderscoreCase(t *testing.T) {
	asserter := assert.New(t)

	var tests = []struct {
		Word   string
		Result string
	}{
		{"userId", "user_id"},
		{"UserId", "user_id"},
		{"user_id", "user_id"},
		{"user-id", "user_id"},
	}

	for _, test := range tests {
		asserter.Equal(lowerUnderScoreCase(test.Word), test.Result)
	}
}

func TestUpperUnderscoreCase(t *testing.T) {
	asserter := assert.New(t)

	var tests = []struct {
		Word   string
		Result string
	}{
		{"userId", "USER_ID"},
		{"UserId", "USER_ID"},
		{"user_id", "USER_ID"},
		{"user-id", "USER_ID"},
	}

	for _, test := range tests {
		asserter.Equal(upperUnderScoreCase(test.Word), test.Result)
	}
}

func TestLowerHyphenCase(t *testing.T) {
	asserter := assert.New(t)
	var tests = []struct {
		Word   string
		Result string
	}{
		{"userId", "user-id"},
		{"UserId", "user-id"},
		{"user_id", "user-id"},
		{"user-id", "user-id"},
	}

	for _, test := range tests {
		asserter.Equal(lowerHyphenCase(test.Word), test.Result)
	}
}

func TestUpperHyphenCase(t *testing.T) {
	asserter := assert.New(t)

	var tests = []struct {
		Word   string
		Result string
	}{
		{"userId", "USER-ID"},
		{"UserId", "USER-ID"},
		{"user_id", "USER-ID"},
		{"user-id", "USER-ID"},
	}

	for _, test := range tests {
		asserter.Equal(upperHyphenCase(test.Word), test.Result)
	}

}
