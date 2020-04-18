package shell

import (
	"bytes"
	"fmt"
	"image"
	"strconv"

	fb "github.com/d3ce1t/areyouin-server/facebook"
	"github.com/d3ce1t/areyouin-server/model"
	"github.com/d3ce1t/areyouin-server/utils"
)

// reset_picture
type resetPictureCmd struct {
}

func (c *resetPictureCmd) Exec(shell *Shell, args []string) {

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
	pictureBytes, err = utils.ResizeImage(originalImage, model.UserPictureMaxWidth)
	manageShellError(err)

	// Change image
	err = shell.model.Accounts.ChangeProfilePicture(userAccount, pictureBytes)
	manageShellError(err)

	fmt.Fprintf(shell, "Picture size %v bytes\n", len(pictureBytes))
}
