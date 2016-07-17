package model

import (
  "errors"
)

var (
  ErrModelAlreadyExist = errors.New("cannot register model because it already exists")
  ErrModelNotFound = errors.New("model not found")
  ErrImageOutOfBounds = errors.New("image is out of bounds")
)
