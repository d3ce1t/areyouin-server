package model

import (
  "bytes"
  core "peeple/areyouin/common"
  "peeple/areyouin/dao"
  "crypto/sha256"
  "image"
)

func newEventManager(parent *AyiModel, session core.DbSession) *EventManager {
  return &EventManager{parent: parent, dbsession: session}
}

type EventManager struct {
  dbsession core.DbSession
  parent *AyiModel
}

// TODO: Add business rules to control if event can be modified, user how is changing event
// image is allowed to do that, and so on.
func (self *EventManager) ChangeEventPicture(event *core.Event, picture []byte) error {

  if picture != nil && len(picture) != 0 {

    // Set event picture

    // Compute digest for picture
    digest := sha256.Sum256(picture)

    corePicture := &core.Picture{
      RawData: picture,
      Digest:  digest[:],
    }

    // Save event picture
    if err := self.saveEventPicture(event.EventId, corePicture); err != nil {
      return err
    }

    event.PictureDigest = corePicture.Digest

  } else {

    // Remove event picture

    if err := self.removeEventPicture(event.EventId); err != nil {
      return err
    }

    event.PictureDigest = nil
  }

  return nil
}

func (self *EventManager) saveEventPicture(event_id int64, picture *core.Picture) error {

	// Decode image
	srcImage, _, err := image.Decode(bytes.NewReader(picture.RawData))
	if err != nil {
		return err
	}

	// Check image size is inside bounds
	if srcImage.Bounds().Dx() > core.EVENT_PICTURE_MAX_WIDTH || srcImage.Bounds().Dy() > core.EVENT_PICTURE_MAX_HEIGHT {
		return ErrImageOutOfBounds
	}

	// Create thumbnails
	thumbnails, err := core.CreateThumbnails(srcImage, EVENT_THUMBNAIL, self.parent.supportedDpi)
	if err != nil {
		return err
	}

	// Save thumbnails
	thumbDAO := dao.NewThumbnailDAO(self.dbsession)
	err = thumbDAO.Insert(event_id, picture.Digest, thumbnails)
	if err != nil {
		return err
	}

	// Save event picture (always does it after thumbnails)
	eventDAO := dao.NewEventDAO(self.dbsession)
	err = eventDAO.SetEventPicture(event_id, picture)
	if err != nil {
		return err
	}

	return nil
}

func (self *EventManager) removeEventPicture(event_id int64) error {

	// Remove event picture
	eventDAO := dao.NewEventDAO(self.dbsession)
	err := eventDAO.SetEventPicture(event_id, &core.Picture{RawData: nil, Digest:  nil})
	if err != nil {
		return err
	}

	// Remove thumbnails
	thumbDAO := dao.NewThumbnailDAO(self.dbsession)
	err = thumbDAO.Remove(event_id)
	if err != nil {
		return err
	}

	return nil
}
