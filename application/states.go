package application

// list of states for the state machine
var (
	Draft     = State{Name: "draft"}
	InReview  = State{Name: "in_review"}
	Approved  = State{Name: "approved"}
	Published = State{Name: "published"}
)
