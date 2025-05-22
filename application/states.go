package application

// list of states for the state machine
var (
	Draft     = State{Name: "DRAFT"}
	InReview  = State{Name: "IN_REVIEW"}
	Approved  = State{Name: "APPROVED"}
	Published = State{Name: "PUBLISHED"}
)
