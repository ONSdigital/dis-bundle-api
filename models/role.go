package models

type Role string

const (
	RoleDatasetsPreviewer Role = "datasets-previewer"
)

func ValidateRole(role Role) bool {
	switch role {
	case RoleDatasetsPreviewer:
		return true
	default:
		return false
	}
}

func (r Role) String() string {
	return string(r)
}
