package model

import "time"

const (
	descriptionMinLength  = 15
	descriptionMaxLength  = 500
	eventPictureMaxWidth  = 1280
	eventPictureMaxHeight = 720

	startDateMinDiff = 30 * time.Minute     // 30 minutes
	startDateMaxDiff = 365 * 24 * time.Hour // 1 year
	endDateMinDiff   = 30 * time.Minute     // 30 minutes (from start date)
	endDateMaxDiff   = 7 * 24 * time.Hour   // 1 week (from start date)
)

// DateOption enum
type DateOption int

// DateOption values
const (
	MinimumStartDate = DateOption(0)
	MaximumStartDate = DateOption(1)
	MinimumEndDate   = DateOption(2)
	MaximumEndDate   = DateOption(3)
)

func GetDateOption(option DateOption, fromDate time.Time) time.Time {

	dateMinute := fromDate.Truncate(time.Minute)

	switch option {
	case MinimumStartDate:
		return dateMinute.Add(startDateMinDiff)
	case MaximumStartDate:
		return dateMinute.Add(startDateMaxDiff)
	case MinimumEndDate:
		return dateMinute.Add(endDateMinDiff)
	case MaximumEndDate:
		return dateMinute.Add(endDateMaxDiff)
	}

	return time.Time{}
}

func IsValidStartDate(startDate time.Time, referenceDate time.Time) bool {

	if startDate.Before(GetDateOption(MinimumStartDate, referenceDate)) ||
		startDate.After(GetDateOption(MaximumStartDate, referenceDate)) {
		return false
	}

	return true
}

func IsValidEndDate(endDate time.Time, referenceDate time.Time) bool {

	if endDate.Before(GetDateOption(MinimumEndDate, referenceDate)) ||
		endDate.After(GetDateOption(MaximumEndDate, referenceDate)) {
		return false
	}

	return true
}

func IsValidDescription(description string) bool {
	if description == "" || len(description) < descriptionMinLength ||
		len(description) > descriptionMaxLength {
		return false
	}
	return true
}

func IsValidName(name string) bool {
	if len(name) < UserNameMinLength || len(name) > UserNameMaxLength {
		return false
	}
	return true
}
