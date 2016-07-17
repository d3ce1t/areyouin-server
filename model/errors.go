package model

import (
  "errors"
)

var (
  ErrModelAlreadyExist = errors.New("cannot register model because it already exists")
  ErrModelNotFound = errors.New("model not found")
  ErrImageOutOfBounds = errors.New("image is out of bounds")
  ErrInvalidUserOrPassword = errors.New("invalid user or password")
  ErrModelInconsistency = errors.New("Model has an inconsistency that requires admin to fix it")
)
