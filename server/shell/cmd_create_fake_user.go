package shell

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"net/http"
	"peeple/areyouin/facebook"
	"peeple/areyouin/model"
	"peeple/areyouin/utils"
	"unicode"
)

type createFakeUserCmd struct {
}

func (c *createFakeUserCmd) Exec(shell *Shell, args []string) {

	var password string
	var linkToFacebook bool

	cmd := flag.NewFlagSet(args[0], flag.ExitOnError)
	cmd.SetOutput(shell)
	cmd.Usage = func() {
		fmt.Fprintf(shell, "Usage of %s:\n", args[0])
		cmd.PrintDefaults()
	}

	cmd.StringVar(&password, "password", "12345", "User password")
	cmd.BoolVar(&linkToFacebook, "facebook", false, "Link to Facebook")

	cmd.Parse(args[1:])

	if cmd.NArg() > 0 {
		cmd.Usage()
		return
	}

	fakeUser, err := c.getRandomFakeUser()
	manageShellError(err)
	fakeUser.password = password

	var user *model.UserAccount

	if linkToFacebook {

		// Create new user account linked to FB

		fbUser, err := facebook.CreateTestUser(fakeUser.name, true)
		manageShellError(err)

		fakeUser.email = fbUser.Email
		user, err = shell.model.Accounts.CreateUserAccount(fakeUser.name, fakeUser.email, fakeUser.password, "", fbUser.Id, fbUser.AccessToken)
		manageShellError(err)

	} else {

		// Create new user account

		user, err = shell.model.Accounts.CreateUserAccount(fakeUser.name, fakeUser.email, fakeUser.password, "", "", "")
		manageShellError(err)
	}

	// Success
	fmt.Fprint(shell, "Account created successfully\n")
	fmt.Fprintf(shell, "Name: %v\nEmail: %v\nPassword: %v\n",
		fakeUser.name, fakeUser.email, fakeUser.password)

	err = shell.model.Accounts.ChangeProfilePicture(user, fakeUser.picture)
	manageShellError(err)

	fmt.Fprintf(shell, "Profile picture changed\n")
}

type FakeUser struct {
	name     string
	email    string
	password string
	picture  []byte
}

func (c *createFakeUserCmd) getRandomFakeUser() (*FakeUser, error) {

	// Request random user data
	resp, err := http.Get("https://randomuser.me/api/")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	jsonData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Decode data
	var f map[string]interface{}
	err = json.Unmarshal(jsonData, &f)
	if err != nil {
		return nil, err
	}

	m := f["results"].([]interface{})[0].(map[string]interface{})
	firstName := []rune(m["name"].(map[string]interface{})["first"].(string))
	lastName := []rune(m["name"].(map[string]interface{})["last"].(string))
	firstName[0] = unicode.ToUpper(firstName[0])
	lastName[0] = unicode.ToUpper(lastName[0])
	name := string(firstName) + " " + string(lastName)
	email := m["email"].(string)
	password := string([]rune(m["login"].(map[string]interface{})["password"].(string)))

	pictures := m["picture"].(map[string]interface{})
	pictureURL := pictures["large"].(string)

	// Download profile picture
	resp, err = http.Get(pictureURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	pictureBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Decode image
	originalImage, _, err := image.Decode(bytes.NewReader(pictureBytes))
	if err != nil {
		return nil, err
	}

	// Resize image to 512xauto
	pictureBytes, err = utils.ResizeImage(originalImage, model.PROFILE_PICTURE_MAX_WIDTH)
	if err != nil {
		return nil, err
	}

	return &FakeUser{
		name:     name,
		email:    email,
		password: password,
		picture:  pictureBytes,
	}, nil
}
