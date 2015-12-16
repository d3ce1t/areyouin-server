package dao

import (
	core "areyouin/common"
	"testing"
)

// Test inserting invalid users and valid users
func TestInsert1(t *testing.T) {

	dao := NewUserDAO(session)
	core.ClearUserAccounts(session)

	var tests = []struct {
		user *core.UserAccount
		want bool
	}{
		{core.NewUserAccount(0, "User 1", "user1@foo.com", "", "", "FBID0", "FBTOKEN0"), false},                     // Invalid ID
		{core.NewUserAccount(15919019823465493, "", "user1@foo.com", "", "", "FBID0", "FBTOKEN0"), false},           // Invalid name
		{core.NewUserAccount(15919019823465493, "em", "user1@foo.com", "", "", "FBID0", "FBTOKEN0"), false},         // Invalid name (less than 3 chars)
		{core.NewUserAccount(15919019823465493, "User 1", "", "", "", "FBID0", "FBTOKEN0"), false},                  // Invalid e-mail
		{core.NewUserAccount(15919019823465493, "User 1", "em@ail", "", "", "FBID0", "FBTOKEN0"), false},            // Invalid e-mail (no match format)
		{core.NewUserAccount(15919019823465493, "User 1", "user1@foo.com", "", "", "", ""), false},                  // There isn't credentials
		{core.NewUserAccount(15919019823465493, "User 1", "user1@foo.com", "123", "", "", ""), false},               // Invalid password (less than 5 chars)
		{core.NewUserAccount(15919019823465493, "User 1", "user1@foo.com", "12345", "", "", ""), true},              // Valid user with e-mail credentials
		{core.NewUserAccount(15918606474806289, "User 2", "user2@foo.com", "", "", "FBID2", ""), false},             // There is facebook id but not token
		{core.NewUserAccount(15918606474806289, "User 2", "user2@foo.com", "", "", "FBID2", "FBTOKEN2"), true},      // Valid user with Facebook credentials
		{core.NewUserAccount(15918606642578451, "User 3", "user3@foo.com", "12345", "", "FBID3", "FBTOKEN3"), true}, // Valid user with both credentials
	}

	for i, test := range tests {
		if applied, err := dao.Insert(test.user); applied != test.want {
			t.Fatal("Failed at test", i, "with error", err)
		}
	}
}

// Test inserting users with same ID, e-mail and Facebook ID
func TestInsert2(t *testing.T) {

	dao := NewUserDAO(session)

	var tests = []struct {
		user *core.UserAccount
		want bool
	}{
		{core.NewUserAccount(15918606642578453, "User 4", "user4@foo.com", "12345", "", "", ""), true},          // Valid user with e-mail credentials
		{core.NewUserAccount(15918606642578453, "User 5", "user5@foo.com", "12345", "", "", ""), false},         // Same ID as user 4
		{core.NewUserAccount(15918606642578452, "User 5", "user4@foo.com", "12345", "", "", ""), false},         // Same e-mail as user 4
		{core.NewUserAccount(15918606642578452, "User 5", "user5@foo.com", "12345", "", "", ""), true},          // Valid user with e-mail credentials
		{core.NewUserAccount(15918606642578453, "User 6", "user6@foo.com", "", "", "FBID6", "FBTOKEN6"), false}, // Same ID as user 4
		{core.NewUserAccount(15918606642578452, "User 6", "user6@foo.com", "", "", "FBID6", "FBTOKEN6"), false}, // Same ID as user 5
		{core.NewUserAccount(15918606642578450, "User 6", "user4@foo.com", "", "", "FBID6", "FBTOKEN6"), false}, // Same e-mail as user 4
		{core.NewUserAccount(15918606642578450, "User 6", "user5@foo.com", "", "", "FBID6", "FBTOKEN6"), false}, // Same e-mail as user 5
		{core.NewUserAccount(15918606642578450, "User 6", "user6@foo.com", "", "", "FBID6", "FBTOKEN6"), true},  // Valid user with Facebook credentials
		{core.NewUserAccount(15919019823465492, "User 7", "user7@foo.com", "", "", "FBID6", "FBTOKEN7"), false}, // Same Facebook ID aser user 6
		{core.NewUserAccount(15919019823465492, "User 7", "user7@foo.com", "", "", "FBID7", "FBTOKEN6"), true},  // Valid user with Facebook credentials (but same token as user6)
	}

	for i, test := range tests {
		if applied, err := dao.Insert(test.user); applied != test.want {
			t.Fatal("Failed at test", i, "with error", err)
		}
	}
}

// Test if e-mail is removed from user_email_credentials when cannot insert into
// user_facebook_credentials
func TestInsert3(t *testing.T) {

	dao := NewUserDAO(session)

	var tests = []struct {
		user *core.UserAccount
		want bool
	}{
		{core.NewUserAccount(15919019823465494, "User 8", "user8@foo.com", "", "", "FBID3", "FBTOKEN8"), false}, // Same Facebook ID as user 3
		{core.NewUserAccount(15919019823465494, "User 8", "user8@foo.com", "", "", "FBID8", "FBTOKEN8"), true},  // If previous test didn't remove e-mail from user_email_credentials, this will fail
	}

	for i, test := range tests {
		if applied, err := dao.Insert(test.user); applied != test.want {
			t.Fatal("Failed at test", i, "with error", err)
		}
	}
}

// Test if e-mail and facebook are removed from user_email_credentials and
// user_facebook_credentials when cannot insert into user_account
func TestInsert4(t *testing.T) {

	dao := NewUserDAO(session)

	var tests = []struct {
		user *core.UserAccount
		want bool
	}{
		{core.NewUserAccount(15919019823465493, "User 9", "user9@foo.com", "", "", "FBID9", "FBTOKEN9"), false}, // Same ID as user 1
		{core.NewUserAccount(15919019823465550, "User 9", "user9@foo.com", "", "", "FBID9", "FBTOKEN9"), true},  // If previous test didn't remove e-mail and facebook, this will fail
	}

	for i, test := range tests {
		if applied, err := dao.Insert(test.user); applied != test.want {
			t.Fatal("Failed at test", i, "with error", err)
		}
	}
}

func TestGetIDByEmail(t *testing.T) {

	dao := NewUserDAO(session)

	var tests = []struct {
		email string
		want  bool
	}{
		{"user1@foo.com", true},
		{"user2@foo.com", true},
		{"user1@bar.com", false},
		{"user3@foo.com", true},
		{"user4@foo.com", true},
		{"user5@foo.com", true},
		{"user6@foo.com", true},
		{"user7@foo.com", true},
		{"user8@foo.com", true},
		{"user9@foo.com", true},
		{"user10@foo.com", false},
	}

	for i, test := range tests {
		if user_id := dao.GetIDByEmail(test.email); (user_id != 0) != test.want {
			t.Fatal("Failed at test", i, user_id)
		}
	}
}

func TestDelete(t *testing.T) {

	dao := NewUserDAO(session)

	var tests = []struct {
		user *core.UserAccount
		want bool
	}{
		{core.NewUserAccount(15919019823465493, "User 1", "user1@foo.com", "12345", "", "", ""), true},              // Remove existing user
		{core.NewUserAccount(15918606474806289, "User 2", "user2@foo.com", "", "", "FBID2", "FBTOKEN2"), true},      // Remove existing user
		{core.NewUserAccount(15918606642578451, "User 3", "user3@foo.com", "12345", "", "FBID3", "FBTOKEN3"), true}, // Remove existing user
		{core.NewUserAccount(15918606642578453, "User 4", "user4@foo.com", "12345", "", "", ""), true},              // Remove existing user
		{core.NewUserAccount(15918606642578452, "User 5", "user5@foo.com", "12345", "", "", ""), true},              // Remove existing user
		{core.NewUserAccount(15918606642578450, "User 6", "user6@foo.com", "", "", "FBID6", "FBTOKEN6"), true},      // Remove existing user
		{core.NewUserAccount(15919019823465492, "User 7", "user7@foo.com", "", "", "FBID7", "FBTOKEN6"), true},      // Remove existing user
		{core.NewUserAccount(15919019823465494, "User 8", "user8@foo.com", "", "", "FBID8", "FBTOKEN8"), true},      // Remove existing user
		{core.NewUserAccount(15919019823465550, "User 9", "user9@foo.com", "", "", "FBID9", "FBTOKEN9"), true},      // Remove existing user
		{core.NewUserAccount(15919019823465493, "User 1", "user1@foo.com", "12345", "", "", ""), true},              // Remove UNexisting user should return true
	}

	for i, test := range tests {
		if err := dao.Delete(test.user); (err == nil) != test.want {
			t.Fatal("Failed at test", i, "with error", err)
		}
	}
}

func TestUUID(t *testing.T) {

	dao := NewUserDAO(session)
	user := core.NewUserAccount(15919019823465493, "User 1", "user1@foo.com", "12345", "", "", "")

	if ok, err := dao.Insert(user); !ok {
		t.Fatal(err)
	}

	same_user := dao.Load(user.Id)

	if same_user == nil {
		t.FailNow()
	}

	if user.AuthToken.String() != same_user.AuthToken.String() {
		t.Fatal("Auth are different but should be the same")
	}

	if err := dao.Delete(user); err != nil {
		t.Fatal(err)
	}
}
