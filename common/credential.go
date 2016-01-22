package common

type EmailCredential struct {
	Email    string
	Password [32]byte
	Salt     [32]byte
	UserId   uint64
}

type FacebookCredential struct {
	Fbid        string
	Fbtoken     string
	UserId      uint64
	CreatedDate int64
}
