package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/jpeg"
	core "peeple/areyouin/common"
	fb "peeple/areyouin/facebook"
	"strconv"
)

// reset_picture
func (shell *Shell) resetPicture(args []string) {

	user_id, err := strconv.ParseUint(args[1], 10, 64)
	manageShellError(err)

	server := shell.server
	userDAO := server.NewUserDAO()

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
	picture_bytes, err = server.resizeImage(original_image, 512)
	manageShellError(err)

	// Compute digest and prepare image
	digest := sha256.Sum256(picture_bytes)

	picture := &core.Picture{
		RawData: picture_bytes,
		Digest:  digest[:],
	}

	// Save profile Picture
	err = server.saveProfilePicture(user_id, picture)
	manageShellError(err)

	fmt.Fprintf(shell.io, "Picture size %v bytes\n", len(picture_bytes))
}
