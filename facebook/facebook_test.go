package facebook

import (
	"testing"
)

func TestCreateAndDeleteTestUser(t *testing.T) {

	user, err := CreateTestUser("Test User One", true)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := DeleteTestUser(user.Id)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.FailNow()
	}

}
