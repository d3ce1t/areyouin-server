package cqldao

import (
	"peeple/areyouin/api"
	"peeple/areyouin/utils"
	"testing"
	"time"
)

func TestTimeLineDAO_Clean(t *testing.T) {
	d1 := NewTimeLineDAO(session)
	d1.DeleteAll()
}

func TestTimeLineDAO_Insert(t *testing.T) {

	rt := time.Date(2016, 10, 10, 12, 0, 0, 0, time.UTC)

	var tests = []struct {
		eventID int64
		endDate time.Time
	}{
		{1, rt},
		{2, rt.Add(2 * time.Hour)},
		{3, rt.Add(4 * time.Hour)},
		{4, rt.Add(6 * time.Hour)},
		{5, rt.Add(8 * time.Hour)},
		{6, rt.AddDate(1, 0, 0)},
	}

	d := NewTimeLineDAO(session)

	for _, test := range tests {

		dto := &api.TimeLineEntryDTO{
			EventID:  test.eventID,
			Position: test.endDate,
		}

		if err := d.Insert(dto); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFindAllFrom(t *testing.T) {

	rt := time.Date(2016, 10, 10, 12, 0, 0, 0, time.UTC)

	var tests = []struct {
		eventID int64
		endDate time.Time
	}{
		{1, rt},
		{2, rt.Add(2 * time.Hour)},
		{3, rt.Add(4 * time.Hour)},
		{4, rt.Add(6 * time.Hour)},
		{5, rt.Add(8 * time.Hour)},
		{6, rt.AddDate(1, 0, 0)},
	}

	d := NewTimeLineDAO(session)

	results, err := d.FindAllFrom(utils.TimeToMillis(rt))
	if err != nil {
		t.Fatal(err)
	}

	for i, item := range results {

		test := tests[i]

		if item.EventID != test.eventID || !item.Position.Equal(test.endDate) {
			t.Fatalf("TestFindAllFrom: %v) Read back different items than inserted in previous testing (item.ID: %v, item.position: %v, test.ID: %v, endDate: %v)",
				i, item.EventID, item.Position, test.eventID, test.endDate)
		}
	}
}

func TestReplace(t *testing.T) {

	rt := time.Date(2016, 10, 10, 12, 0, 0, 0, time.UTC)

	var tests = []struct {
		eventID    int64
		endDate    time.Time
		newEndDate time.Time
	}{
		{1, rt, rt.Add(1 * time.Hour)},
		{2, rt.Add(2 * time.Hour), rt.Add(3 * time.Hour)},
		{3, rt.Add(4 * time.Hour), rt.Add(4 * time.Hour)},
		{4, rt.Add(6 * time.Hour), rt.Add(5 * time.Hour)},
		{5, rt.Add(8 * time.Hour), rt.Add(6 * time.Hour)},
		{6, rt.AddDate(1, 0, 0), rt.Add(7 * time.Hour)},
	}

	d := NewTimeLineDAO(session)

	// Execute test
	for _, test := range tests {

		oldDto := &api.TimeLineEntryDTO{
			EventID:  test.eventID,
			Position: test.endDate,
		}

		newDto := &api.TimeLineEntryDTO{
			EventID:  test.eventID,
			Position: test.newEndDate,
		}

		if err := d.Replace(oldDto, newDto); err != nil {
			t.Fatal(err)
		}
	}

	// Check results
	results, err := d.FindAllFrom(utils.TimeToMillis(rt))
	if err != nil {
		t.Fatal(err)
	}

	for i, item := range results {

		test := tests[i]

		if item.EventID != test.eventID || item.Position != test.newEndDate {
			t.Fatal("Read back different items than replaced")
		}
	}
}

func TestDelete(t *testing.T) {

	rt := time.Date(2016, 10, 10, 12, 0, 0, 0, time.UTC)

	var tests = []struct {
		eventID int64
		endDate time.Time
	}{
		{1, rt.Add(1 * time.Hour)},
		{2, rt.Add(3 * time.Hour)},
		{3, rt.Add(4 * time.Hour)},
		{4, rt.Add(5 * time.Hour)},
		{5, rt.Add(6 * time.Hour)},
		{6, rt.Add(7 * time.Hour)},
	}

	d := NewTimeLineDAO(session)

	for _, test := range tests {

		dto := &api.TimeLineEntryDTO{
			EventID:  test.eventID,
			Position: test.endDate,
		}

		if err := d.Delete(dto); err != nil {
			t.Fatal("Delete error")
		}
	}

	results, err := d.FindAllFrom(utils.TimeToMillis(rt))
	if err != nil {
		t.Fatal("Delete error: Could not check deletion")
	}

	if len(results) > 0 {
		t.Fatal("Delete error: Count mismatch")
	}
}
