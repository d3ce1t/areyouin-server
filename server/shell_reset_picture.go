package main

import _ "image/jpeg"

// reset_picture
func (shell *Shell) resetPicture(args []string) {

	/*user_id, err := strconv.ParseInt(args[1], 10, 64)
	manageShellError(err)

	server := shell.server
	userDAO := dao.NewUserDAO(server.DbSession)

	// Load user
	user_account, err := userDAO.Load(user_id)
	manageShellError(err)

	// Get profile picture
	fbsession := fb.NewSession(user_account.Fbtoken)
	picture_bytes, err := fb.GetProfilePicture(fbsession)
	manageShellError(err)

	// Decode image
	original_image, _, err := image.Decode(bytes.NewReader(picture_bytes))
	manageShellError(err)

	// Resize image to 512x512
	picture_bytes, err = core.ResizeImage(original_image, core.PROFILE_PICTURE_MAX_WIDTH)
	manageShellError(err)

	// Change image
	err = server.Model.Accounts.ChangeProfilePicture(user_account, picture_bytes)
	manageShellError(err)

	fmt.Fprintf(shell.io, "Picture size %v bytes\n", len(picture_bytes))*/
}
