package facebook

type FacebookAccount struct {
	Id          string
	Name        string
	Email       string
	Password    string
	AccessToken string
}

type Friend struct {
	Id   string
	Name string
}
