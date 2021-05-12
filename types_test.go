package raml

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTypeInType(t *testing.T) {
	apiDef := new(APIDefinition)
	Convey("Type in type's properties", t, func() {
		err := ParseFile("./samples/types.raml", apiDef)
		So(err, ShouldBeNil)

		action, ok := apiDef.Types["Action"]
		So(ok, ShouldBeTrue)

		name := action.GetProperty("name")
		So(name.Type, ShouldEqual, "string")

		recurring := action.GetProperty("recurring")
		So(recurring.TypeString(), ShouldEqual, "Actionrecurring")

		// check the inline type
		ar, ok := apiDef.Types["Actionrecurring"]
		So(ok, ShouldBeTrue)

		// Must work via .GetPropert
		period := ar.GetProperty("period")
		So(period.TypeString(), ShouldEqual, "integer")

		// Also must work via .ToProperty
		var prop Property
		for k, p := range action.Properties {
			if k == "recurring" {
				prop = ToProperty(k, p)
				break
			}
		}
		So(prop.TypeString(), ShouldEqual, "Actionrecurring")

		// test for the recursive type
		_, ok = apiDef.Types["Actionrecurringcombo"]
		So(ok, ShouldBeTrue)

		combo := ar.GetProperty("combo")
		So(combo.TypeString(), ShouldEqual, "Actionrecurringcombo")

		// check the items with properties
		coinInputs := action.GetProperty("coininputs")
		So(coinInputs.Type, ShouldEqual, "array")
		So(coinInputs.Items.Type, ShouldEqual, "ActioncoininputsItem")

		_, ok = apiDef.Types["ActioncoininputsItem"]
		So(ok, ShouldBeTrue)

		// check the items with Type and format
		coinTipes := action.GetProperty("coinTipes")
		So(coinTipes.Type, ShouldEqual, "array")
		So(coinTipes.Items.Type, ShouldEqual, "number")
		So(coinTipes.Items.Format, ShouldEqual, "double")

		// check the items with plain type
		coinTipesPlain := action.GetProperty("coinTipesPlain")
		So(coinTipesPlain.Type, ShouldEqual, "array")
		So(coinTipesPlain.Items.Type, ShouldEqual, "string")
	})
}
