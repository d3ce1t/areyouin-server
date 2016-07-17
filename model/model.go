package model

import (
  core "peeple/areyouin/common"
  "math"
  "sync"
)

const (
  ALL_CONTACTS_GROUP         = 0   // Id for the main friend group of a user
  THUMBNAIL_MDPI_SIZE        = 50  // 50 px
  EVENT_THUMBNAIL            = 100 // 100 px
)

var (
  registeredModels *ModelsMap
  mutex sync.RWMutex
)

func init() {
  registeredModels = newModelsMap()
}

// Creates a new model with the given key for later retrieval. If model exist panic
func New(session core.DbSession, key string) *AyiModel {
  defer mutex.RUnlock()
	mutex.RLock()

  model, ok := registeredModels.Get(key)

  if !ok {
    model = &AyiModel{
      supportedDpi: []int32{core.IMAGE_MDPI, core.IMAGE_HDPI, core.IMAGE_XHDPI,
        core.IMAGE_XXHDPI, core.IMAGE_XXXHDPI},
    }
    model.Accounts = newAccountManager(model, session)
    model.Events = newEventManager(model, session)
  } else {
    panic(ErrModelAlreadyExist)
  }

  return model
}

// Gets an already existing model and panic if model does not exist
func Get(key string) *AyiModel {
  defer mutex.RUnlock()
	mutex.RLock()
  if model, ok := registeredModels.Get(key); ok {
    return model
  } else {
    panic(ErrModelNotFound)
  }
}

type AyiModel struct {
  dbHost string
  dbName string
  supportedDpi []int32
  Accounts *AccountManager
  Events *EventManager
}

func (self *AyiModel) GetClosestDpi(reqDpi int32) int32 {

	if reqDpi <= core.IMAGE_MDPI {
		return core.IMAGE_MDPI
	} else if reqDpi >= core.IMAGE_XXXHDPI {
		return core.IMAGE_XXXHDPI
	}

	min_dist := math.MaxFloat32
	dpi_index := 0

	for i, dpi := range self.supportedDpi {
		dist := math.Abs(float64(reqDpi - dpi))
		if dist < min_dist {
			min_dist = dist
			dpi_index = i
		}
	}

	if self.supportedDpi[dpi_index] < reqDpi {
		dpi_index++
	}

	return self.supportedDpi[dpi_index]
}
