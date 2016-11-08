package cqldao

import (
	"peeple/areyouin/utils"
	"testing"
)

func TestTimeLineDAO_Insert(t *testing.T) {

	d := NewTimeLineDAO(session)
	origin := utils.CreateDate(2016, 11, 3, 12, 0)

	for _, dto := range generateTimelineEntries(origin, 50) {
		if err := d.Insert(dto); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTimeLineDAO_FindAllForward(t *testing.T) {

	d := NewTimeLineDAO(session)

	// Clear
	if err := d.DeleteAll(); err != nil {
		t.Fatal(err)
	}

	// Insert
	origin := utils.CreateDate(2016, 11, 3, 12, 0)
	entries := generateTimelineEntries(origin, 50)
	for _, dto := range entries {
		if err := d.Insert(dto); err != nil {
			t.Fatal(err)
		}
	}

	// Read
	results, err := d.FindAllForward(origin)
	if err != nil {
		t.Fatal(err)
	}

	// Check
	for i, item := range results {
		test := entries[i]
		if item.EventID != test.EventID || !item.Position.Equal(test.Position) {
			t.Fatalf("Read back different items than inserted")
		}
	}
}

func TestTimeLineDAO_FindAllBackward(t *testing.T) {

	d := NewTimeLineDAO(session)

	// Clear
	if err := d.DeleteAll(); err != nil {
		t.Fatal(err)
	}

	// Insert
	origin := utils.CreateDate(2016, 11, 3, 12, 0)
	entries := generateTimelineEntries(origin, 50)
	for _, dto := range entries {
		if err := d.Insert(dto); err != nil {
			t.Fatal(err)
		}
	}

	// Read
	results, err := d.FindAllBackward(entries[len(entries)-1].Position)
	if err != nil {
		t.Fatal(err)
	}

	// Check
	for i, result := range results {
		e := entries[len(entries)-i-1]
		if result.EventID != e.EventID || !result.Position.Equal(e.Position) {
			t.Fatalf("Read back different items than inserted")
		}
	}
}

func TestTimeLineDAO_FindAllBetween(t *testing.T) {

	d := NewTimeLineDAO(session)

	// Clear
	if err := d.DeleteAll(); err != nil {
		t.Fatal(err)
	}

	// Insert
	origin := utils.CreateDate(2016, 11, 3, 12, 0)
	entries := generateTimelineEntries(origin, 50)
	for _, dto := range entries {
		if err := d.Insert(dto); err != nil {
			t.Fatal(err)
		}
	}

	// Read
	startIndex := 10
	endIndex := len(entries) - 10
	results, err := d.FindAllBetween(entries[startIndex].Position, entries[endIndex].Position)
	if err != nil {
		t.Fatal(err)
	}

	// Check
	if len(results) != (endIndex - startIndex + 1) {
		t.Fatal("Count mismatch")
	}

	for i, entry := range entries[startIndex : endIndex+1] {
		e := results[i]
		if entry.EventID != e.EventID || !entry.Position.Equal(e.Position) {
			t.Fatal("Read back different items than inserted")
		}
	}
}

func TestTimeLineDAO_Replace(t *testing.T) {

	d := NewTimeLineDAO(session)

	// Clear
	if err := d.DeleteAll(); err != nil {
		t.Fatal(err)
	}

	// Insert
	origin := utils.CreateDate(2016, 11, 3, 12, 0)
	groundtruth := generateTimelineEntries(origin, 50)
	for _, dto := range groundtruth {
		if err := d.Insert(dto); err != nil {
			t.Fatal(err)
		}
	}

	// Replace
	for i, oldDto := range groundtruth[0 : len(groundtruth)-1] {
		if err := d.Replace(oldDto, groundtruth[i+1]); err != nil {
			t.Fatal(err)
		}
	}

	// Check results

	// Read
	results, err := d.FindAllForward(groundtruth[1].Position)
	if err != nil {
		t.Fatal(err)
	}

	// Check
	for i, item := range results[0 : len(results)-1] {
		test := groundtruth[i+1]
		if item.EventID != test.EventID || !item.Position.Equal(test.Position) {
			t.Fatal("Read back different items than replaced")
		}
	}
}

func TestTimeLineDAO_Delete(t *testing.T) {

	// Clear
	d := NewTimeLineDAO(session)
	if err := d.DeleteAll(); err != nil {
		t.Fatal(err)
	}

	// Insert
	origin := utils.CreateDate(2016, 11, 3, 12, 0)
	entries := generateTimelineEntries(origin, 50)
	for _, dto := range entries {
		if err := d.Insert(dto); err != nil {
			t.Fatal(err)
		}
	}

	// Remove
	for _, dto := range entries {
		if err := d.Delete(dto); err != nil {
			t.Fatal("Delete error")
		}
	}

	// Check
	results, err := d.FindAllForward(origin)
	if err != nil {
		t.Fatal("Delete error: Could not check deletion")
	}

	if len(results) > 0 {
		t.Fatal("Delete error: Count mismatch")
	}
}
