package model

import (
  "errors"
)

var (
  ErrModelAlreadyExist = errors.New("cannot register model because it already exists")
  ErrModelNotFound = errors.New("model not found")
  ErrImageOutOfBounds = errors.New("image is out of bounds")
  ErrInvalidUserOrPassword = errors.New("invalid user or password")
  ErrInvalidAuthor = errors.New("invalid author")
  ErrParticipantsRequired = errors.New("participants required")
  ErrEventOutOfCreationWindow = errors.New("event out of allowed creationg window")
  ErrAuthorDeliveryError = errors.New("event coudn't be delivered to author")
  ErrEventNotWritable = errors.New("event isn't writable")
  ErrModelInconsistency = errors.New("Model has an inconsistency that requires admin to fix it")
)
