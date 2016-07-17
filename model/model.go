package model

import (
  core "peeple/areyouin/common"
  "math"
)

var (
  dbHost string
  dbName string
  supportedDpi []int32
)

const (
  ALL_CONTACTS_GROUP         = 0   // Id for the main friend group of a user
  THUMBNAIL_MDPI_SIZE        = 50  // 50 px
  EVENT_THUMBNAIL            = 100 // 100 px
)

func init() {
  supportedDpi = []int32{core.IMAGE_MDPI, core.IMAGE_HDPI, core.IMAGE_XHDPI,
    core.IMAGE_XXHDPI, core.IMAGE_XXXHDPI}
}

func GetClosestDpi(reqDpi int32) int32 {

	if reqDpi <= core.IMAGE_MDPI {
		return core.IMAGE_MDPI
	} else if reqDpi >= core.IMAGE_XXXHDPI {
		return core.IMAGE_XXXHDPI
	}

	min_dist := math.MaxFloat32
	dpi_index := 0

	for i, dpi := range supportedDpi {
		dist := math.Abs(float64(reqDpi - dpi))
		if dist < min_dist {
			min_dist = dist
			dpi_index = i
		}
	}

	if supportedDpi[dpi_index] < reqDpi {
		dpi_index++
	}

	return supportedDpi[dpi_index]
}
