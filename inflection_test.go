package raml

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInflection(t *testing.T) {
	Convey("Singularize", t, func() {
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
			So(singularize(test.Word), ShouldEqual, test.Result)
		}
	})

	Convey("pluralize", t, func() {
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
			So(pluralize(test.Word), ShouldEqual, test.Result)
		}
	})

	Convey("Lower Camel Case", t, func() {
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
			So(lowerCamelCase(test.Word), ShouldEqual, test.Result)
		}
	})

	Convey("Upper Camel Case", t, func() {
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
			So(upperCamelCase(test.Word), ShouldEqual, test.Result)
		}
	})

	Convey("lower underscore case", t, func() {
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
			So(lowerUnderScoreCase(test.Word), ShouldEqual, test.Result)
		}
	})

	Convey("upper underscore case", t, func() {
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
			So(upperUnderScoreCase(test.Word), ShouldEqual, test.Result)
		}
	})

	Convey("lower hyphen case", t, func() {
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
			So(lowerHyphenCase(test.Word), ShouldEqual, test.Result)
		}
	})

	Convey("upper hypen case", t, func() {
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
			So(upperHyphenCase(test.Word), ShouldEqual, test.Result)
		}
	})

}
