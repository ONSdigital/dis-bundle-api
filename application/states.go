package application

// list of states for the state machine
var Published = State{
	Name:      "PUBLISHED",
	EnterFunc: PublishBundle,
}

var InReview = State{
	Name:      "IN_REVIEW",
	EnterFunc: ReviewBundle,
}

var Approved = State{
	Name:      "APPROVED",
	EnterFunc: ApproveBundle,
}

var Draft = State{
	Name:      "DRAFT",
	EnterFunc: DraftBundle,
}
