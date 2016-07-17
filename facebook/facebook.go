package facebook

import (
	"fmt"
	fb "github.com/huandu/facebook"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	FB_APP_ID     = "888355654618282"
	FB_APP_SECRET = "eac8e246b9e8a8f5a80d722a556f2cec"
	FB_APP_TOKEN  = "888355654618282|raofWCIOKKmKLtY7-Vvjnw5emB4"
)

var App *fb.App

func init() {
	fb.Version = "v2.5"
	App = fb.New(FB_APP_ID, FB_APP_SECRET)
	App.EnableAppsecretProof = true
}

func NewSession(access_token string) *fb.Session {
	return App.Session(access_token)
}

func CheckAccess(id string, session *fb.Session) (*FacebookAccount, error) {

	// Contact Facebook
	res, err := session.Get("/me", fb.Params{
		"fields": "id,name,email",
	})

	if err != nil {
		return nil, ErrFacebookAccessForbidden
	}

	// Get info
	account := &FacebookAccount{}

	if fbid, ok := res["id"]; ok {
		account.Id = fbid.(string)
	} else {
		return nil, ErrMissingFields
	}

	if name, ok := res["name"]; ok {
		account.Name = name.(string)
	} else {
		return nil, ErrMissingFields
	}

	// If there isn't e-mail, then do not give an error
	if email, ok := res["email"]; ok {
		account.Email = email.(string)
	}

	// Check
	if account.Id != id {
		return nil, ErrFacebookAccessForbidden
	}

	return account, nil
}

func CreateTestUser(name string, installed bool) (*FacebookAccount, error) {

	session := App.Session(FB_APP_TOKEN)

	res, err := session.Post("/"+FB_APP_ID+"/accounts", fb.Params{
		"name":        name,
		"installed":   installed,
		"permissions": "public_profile,user_friends,email",
	})

	if err != nil {
		return nil, err
	}

	user := &FacebookAccount{Name: name}

	// Check response fields
	if id, ok := res["id"]; ok {
		user.Id = id.(string)
	} else {
		return nil, ErrMissingFields
	}

	if email, ok := res["email"]; ok {
		user.Email = email.(string)
	} else {
		return nil, ErrMissingFields
	}

	if password, ok := res["password"]; ok {
		user.Password = password.(string)
	} else {
		return nil, ErrMissingFields
	}

	if access_token, ok := res["access_token"]; ok {
		user.AccessToken = access_token.(string)
	} else {
		return nil, ErrMissingFields
	}

	return user, nil
}

func DeleteTestUser(id string) (bool, error) {

	session := App.Session(FB_APP_TOKEN)

	res, err := session.Delete("/"+id, nil)

	if err != nil {
		return false, err
	}

	success, ok := res["success"]

	if !ok {
		return false, ErrMissingFields
	}

	return success.(bool), nil
}

// Connect Facebook and Get Friends
func GetFriends(session *fb.Session) ([]*Friend, error) {

	res, err := session.Get("/me/friends", nil)

	if err != nil {
		return nil, err
	}

	// Manage response
	paging, err := res.Paging(session)

	if err != nil {
		return nil, err
	}

	var friends []*Friend

get_data:
	results := paging.Data()

	for _, res := range results {
		if f, err := parseFriend(res); err == nil {
			friends = append(friends, f)
		} else {
			return nil, err
		}
	}

	if noMore, err := paging.Next(); err != nil {
		return nil, err
	} else if !noMore {
		goto get_data
	}

	return friends, nil
}

func GetProfilePicture(session *fb.Session) ([]byte, error) {

	res, err := session.Get("/me/picture", fb.Params{
		"width":    "618",
		"height":   "618",
		"redirect": "false",
	})

	if err != nil {
		return nil, err
	}

	var data fb.Result

	err = res.DecodeField("data", &data)
	if err != nil {
		return nil, err
	}

	url, ok := data["url"]
	if !ok {
		return nil, ErrMissingFields
	}

	resp, err := http.Get(url.(string))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	picture, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return picture, nil
}

func parseFriend(result fb.Result) (*Friend, error) {

	friend := &Friend{}

	if id, ok := result["id"]; ok {
		friend.Id = id.(string)
	} else {
		return nil, ErrMissingFields
	}

	if name, ok := result["name"]; ok {
		friend.Name = name.(string)
	} else {
		return nil, ErrMissingFields
	}

	return friend, nil
}

func GetErrorMessage(err error) string {

	message := ""

	if err != nil {
		if e, ok := err.(*fb.Error); ok {
			message = fmt.Sprintf("[message:%v] [type:%v] [code:%v] [subcode:%v]",
				e.Message, e.Type, e.Code, e.ErrorSubcode)
		} else {
			message = fmt.Sprintf("%v", err)
		}
	}

	return message
}

func LogError(err error) {
	if err != nil {
		if e, ok := err.(*fb.Error); ok {
			log.Printf("Facebook Error: [message:%v] [type:%v] [code:%v] [subcode:%v]\n",
				e.Message, e.Type, e.Code, e.ErrorSubcode)
		} else {
			log.Println("Facebook Error:", err)
		}
	}
}
