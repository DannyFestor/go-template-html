package domain

type RememberToken struct {
	Selector  string
	Validator string
}

type RememberTokenUse struct {
	UserID UserID
	Replay bool
}
