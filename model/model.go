package model

import (
	"math"
	"sync"

	"github.com/d3ce1t/areyouin-server/api"
	"github.com/d3ce1t/areyouin-server/utils"
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
	sync.RWMutex
	supportedDpi []int32
	dbsession    api.DbSession
	Accounts     *AccountManager
	Events       *EventManager
	Friends      *FriendManager
	initialised  bool
}

// Creates a new model with the given key for later retrieval. If model exist panic
func New(session api.DbSession, key string) *AyiModel {
	defer mutex.RUnlock()
	mutex.RLock()

	model, ok := registeredModels.Get(key)
	if ok {
		panic(ErrModelAlreadyExist)
	}

	model = &AyiModel{
		supportedDpi: []int32{utils.ImageMdpi,
			utils.ImageHdpi,
			utils.ImageXhdpi,
			utils.ImageXxhdpi,
			utils.ImageXxxhdpi},
		dbsession: session,
	}
	model.Accounts = newAccountManager(model, session)
	model.Events = newEventManager(model, session)
	model.Friends = newFriendManager(model, session)
	registeredModels.Put(key, model)

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

// Start init background and start tasks required for this model to work over time
func (m *AyiModel) StartBackgroundTasks() {
	defer m.Unlock()
	m.Lock()
	if !m.initialised {
		m.Events.initBackgroundTasks()
		m.initialised = true
	}
}

func (m *AyiModel) DbSession() api.DbSession {
	return m.dbsession
}

func (self *AyiModel) GetClosestDpi(reqDpi int32) int32 {

	if reqDpi <= utils.ImageMdpi {
		return utils.ImageMdpi
	} else if reqDpi >= utils.ImageXxxhdpi {
		return utils.ImageXxxhdpi
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

func UserMapKeys(m map[int64]*UserAccount) []int64 {
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func ParticipantMapKeys(m map[int64]*Participant) []int64 {
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func FriendKeys(m []*Friend) []int64 {
	keys := make([]int64, 0, len(m))
	for _, f := range m {
		keys = append(keys, f.id)
	}
	return keys
}
