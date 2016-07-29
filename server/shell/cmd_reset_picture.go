package shell

import (
	"bytes"
	"fmt"
	"image"
	fb "peeple/areyouin/facebook"
	"peeple/areyouin/model"
	"peeple/areyouin/utils"
	"strconv"
)

// reset_picture
func resetPicture(shell *Shell, args []string) {

	userID, err := strconv.ParseInt(args[1], 10, 64)
	manageShellError(err)

	// Load user
	userAccount, err := shell.model.Accounts.GetUserAccount(userID)
	manageShellError(err)

	// Get profile picture
	fbsession := fb.NewSession(userAccount.FbToken())
	pictureBytes, err := fb.GetProfilePicture(fbsession)
	manageShellError(err)

	// Decode image
	originalImage, _, err := image.Decode(bytes.NewReader(pictureBytes))
	manageShellError(err)

	// Resize image to 512x512
	pictureBytes, err = utils.ResizeImage(originalImage, model.PROFILE_PICTURE_MAX_WIDTH)
	manageShellError(err)

	// Change image
	err = shell.model.Accounts.ChangeProfilePicture(userAccount, pictureBytes)
	manageShellError(err)

	fmt.Fprintf(shell, "Picture size %v bytes\n", len(pictureBytes))
}
