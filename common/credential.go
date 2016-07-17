package common

type AuthCredential struct {
	UserId int64
	Token string
}

type EmailCredential struct {
	Email    string
	Password [32]byte
	Salt     [32]byte
	UserId   int64
}

type FacebookCredential struct {
	Fbid        string
	Fbtoken     string
	UserId      int64
	CreatedDate int64
}
