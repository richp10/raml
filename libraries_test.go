package raml

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLibraries(t *testing.T) {
	Convey("Libraries", t, func() {
		apiDef := new(APIDefinition)
		err := ParseFile("./samples/simple_with_lib.raml", apiDef)
		So(err, ShouldBeNil)

		Convey("two level library", func() {
			// check Uses
			So(apiDef.Uses, ShouldContainKey, "files")
			So(apiDef.Uses["files"], ShouldEqual, "libraries/files.raml")

			// Check Libraries property
			So(apiDef.Libraries, ShouldContainKey, "files")

			// first level
			files := apiDef.Libraries["files"]
			So(files.Usage, ShouldEqual, "Use to define some basic file-related constructs.")
			So(files.Traits, ShouldContainKey, "drm")
			So(files.Uses, ShouldContainKey, "file-type")
			So(files.ResourceTypes, ShouldContainKey, "file")

			// check trait usage in a resource type
			file := files.ResourceTypes["file"]
			So(file.Get, ShouldNotBeNil)
			So(file.Get.Headers, ShouldContainKey, HTTPHeader("drm-key"))

			// second level
			So(files.Libraries, ShouldContainKey, "file-type")
			fileType := files.Libraries["file-type"]
			So(fileType.Types, ShouldContainKey, "File")
			File := fileType.Types["File"]
			So(len(File.Properties), ShouldEqual, 2)
		})

		Convey("using library's trait in root's definition", func() {
			files := apiDef.Resources["/files"]
			So(files.Get, ShouldNotBeNil)
			So(files.Get.Headers, ShouldContainKey, HTTPHeader("drm-key"))
			So(files.Get.Headers["drm-key"].Required, ShouldBeFalse)
		})

		Convey("proper variable name", func() {
			r := apiDef.Resources["/links"]

			So(r.Post, ShouldNotBeNil)
			So(r.Post.Bodies.Type, ShouldEqual, "files.Link")
		})

	})
}
