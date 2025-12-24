package models

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateRole(t *testing.T) {
	Convey("Given a valid role", t, func() {
		role := RoleDatasetsPreviewer

		Convey("When ValidateRole is called", func() {
			isValid := ValidateRole(role)

			Convey("Then it should return true", func() {
				So(isValid, ShouldBeTrue)
			})
		})
	})

	Convey("Given an invalid role", t, func() {
		role := Role("invalid-role")

		Convey("When ValidateRole is called", func() {
			isValid := ValidateRole(role)

			Convey("Then it should return false", func() {
				So(isValid, ShouldBeFalse)
			})
		})
	})
}
