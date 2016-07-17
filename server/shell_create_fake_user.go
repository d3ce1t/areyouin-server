package main // create_fake_user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"net/http"
	"unicode"
	core "peeple/areyouin/common"
)

func (shell *Shell) createFakeUser(args []string) {

	// Request random user data
	resp, err := http.Get("https://randomuser.me/api/")
	manageShellError(err)

	defer resp.Body.Close()

	json_data, err := ioutil.ReadAll(resp.Body)
	manageShellError(err)

	// Decode data
	var f interface{}
	err = json.Unmarshal(json_data, &f)
	manageShellError(err)

	m := f.(map[string]interface{})["results"].([]interface{})[0].(map[string]interface{})
	m = m["user"].(map[string]interface{})

	firstName := []rune(m["name"].(map[string]interface{})["first"].(string))
	lastName := []rune(m["name"].(map[string]interface{})["last"].(string))
	firstName[0] = unicode.ToUpper(firstName[0])
	lastName[0] = unicode.ToUpper(lastName[0])
	name := string(firstName) + " " + string(lastName)

	email := m["email"].(string)
	password := "12345" //m["password"].(string)

	pictures := m["picture"].(map[string]interface{})
	picture_url := pictures["large"].(string)

	// Download profile picture
	resp, err = http.Get(picture_url)
	manageShellError(err)

	defer resp.Body.Close()
	picture_bytes, err := ioutil.ReadAll(resp.Body)
	manageShellError(err)

	// Decode image
	original_image, _, err := image.Decode(bytes.NewReader(picture_bytes))
	manageShellError(err)

	// Resize image to 512xauto
	picture_bytes, err = core.ResizeImage(original_image, core.PROFILE_PICTURE_MAX_WIDTH)
	manageShellError(err)

	// Create new user account
	user, err := shell.server.Model.Accounts.CreateUserAccount(name, email, password, "", "", "")
	manageShellError(err)

	// Success
	fmt.Fprint(shell.io, "Account created successfully\n")
	fmt.Fprintf(shell.io, "Name: %v\nEmail: %v\nPassword: %v\nPicture: %v\n",
		name, email, password, picture_url)

	err = shell.server.Model.Accounts.ChangeProfilePicture(user, picture_bytes)
	manageShellError(err)

	fmt.Fprintf(shell.io, "Profile picture changed\n")
}
