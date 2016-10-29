package model

import (
	"math"
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
	"sync"
)

const (
	allContactsGroup   = 0   // Id for the main friend group of a user
	userThumbnailSize  = 50  // 50 px
	eventThumbnailSize = 100 // 100 px
)

var (
	registeredModels *ModelsMap
	mutex            sync.RWMutex
)

func init() {
	registeredModels = newModelsMap()
}

type AyiModel struct {
	supportedDpi []int32
	dbsession    api.DbSession
	Accounts     *AccountManager
	Events       *EventManager
	Friends      *FriendManager
}

// Creates a new model with the given key for later retrieval. If model exist panic
func New(session api.DbSession, key string) *AyiModel {
	defer mutex.RUnlock()
	mutex.RLock()

	model, ok := registeredModels.Get(key)

	if !ok {
		model = &AyiModel{
			supportedDpi: []int32{utils.IMAGE_MDPI,
				utils.IMAGE_HDPI,
				utils.IMAGE_XHDPI,
				utils.IMAGE_XXHDPI,
				utils.IMAGE_XXXHDPI},
			dbsession: session,
		}
		model.Accounts = newAccountManager(model, session)
		model.Events = newEventManager(model, session)
		model.Friends = newFriendManager(model, session)
		registeredModels.Put(key, model)
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

func getUserMapKeys(m map[int64]*UserAccount) []int64 {
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getParticipantMapKeys(m map[int64]*Participant) []int64 {
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (m *AyiModel) DbSession() api.DbSession {
	return m.dbsession
}

func (self *AyiModel) GetClosestDpi(reqDpi int32) int32 {

	if reqDpi <= utils.IMAGE_MDPI {
		return utils.IMAGE_MDPI
	} else if reqDpi >= utils.IMAGE_XXXHDPI {
		return utils.IMAGE_XXXHDPI
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
