package cqldao

/*import (
	core "github.com/d3ce1t/areyouin-server/common"
	"testing"
	"time"
)

// Test inserting invalid users and valid users
func TestInsert1(t *testing.T) {

	dao := NewUserDAO(session)
	core.DeleteFakeusers(dao)
	time.Sleep(2 * time.Second)

	var tests = []struct {
		user *core.UserAccount
		want bool
	}{
		{core.NewUserAccount(0, "User 1", "user1@foo.com", "", "", "FBID0", "FBTOKEN0"), false},                     // Invalid ID
		{core.NewUserAccount(15918606474806281, "", "user1@foo.com", "", "", "FBID0", "FBTOKEN0"), false},           // Invalid name
		{core.NewUserAccount(15918606474806281, "em", "user1@foo.com", "", "", "FBID0", "FBTOKEN0"), false},         // Invalid name (less than 3 chars)
		{core.NewUserAccount(15918606474806281, "User 1", "", "", "", "FBID0", "FBTOKEN0"), false},                  // Invalid e-mail
		{core.NewUserAccount(15918606474806281, "User 1", "em@ail", "", "", "FBID0", "FBTOKEN0"), false},            // Invalid e-mail (no match format)
		{core.NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "", "", "", ""), false},                  // There isn't credentials
		{core.NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "123", "", "", ""), false},               // Invalid password (less than 5 chars)
		{core.NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "12345", "", "", ""), true},              // Valid user with e-mail credentials
		{core.NewUserAccount(15918606642578452, "User 2", "user2@foo.com", "", "", "FBID2", ""), false},             // There is facebook id but not token
		{core.NewUserAccount(15918606642578452, "User 2", "user2@foo.com", "", "", "FBID2", "FBTOKEN2"), true},      // Valid user with Facebook credentials
		{core.NewUserAccount(15918606642578453, "User 3", "user3@foo.com", "12345", "", "FBID3", "FBTOKEN3"), true}, // Valid user with both credentials
	}

	for i, test := range tests {
		if err := dao.Insert(test.user); (err == nil) != test.want {
			t.Fatal("Failed at test", i, test.user.Id, "with error", err)
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
		{core.NewUserAccount(15918606642578454, "User 4", "user4@foo.com", "12345", "", "", ""), true},          // Valid user with e-mail credentials
		{core.NewUserAccount(15918606642578454, "User 5", "user5@foo.com", "12345", "", "", ""), false},         // Same ID as user 4
		{core.NewUserAccount(15919019823465485, "User 5", "user4@foo.com", "12345", "", "", ""), false},         // Same e-mail as user 4
		{core.NewUserAccount(15919019823465485, "User 5", "user5@foo.com", "12345", "", "", ""), true},          // Valid user with e-mail credentials
		{core.NewUserAccount(15918606642578454, "User 6", "user6@foo.com", "", "", "FBID6", "FBTOKEN6"), false}, // Same ID as user 4
		{core.NewUserAccount(15919019823465485, "User 6", "user6@foo.com", "", "", "FBID6", "FBTOKEN6"), false}, // Same ID as user 5
		{core.NewUserAccount(15919019823465496, "User 6", "user4@foo.com", "", "", "FBID6", "FBTOKEN6"), false}, // Same e-mail as user 4
		{core.NewUserAccount(15919019823465496, "User 6", "user5@foo.com", "", "", "FBID6", "FBTOKEN6"), false}, // Same e-mail as user 5
		{core.NewUserAccount(15919019823465496, "User 6", "user6@foo.com", "", "", "FBID6", "FBTOKEN6"), true},  // Valid user with Facebook credentials
		{core.NewUserAccount(15919019823465497, "User 7", "user7@foo.com", "", "", "FBID6", "FBTOKEN7"), false}, // Same Facebook ID aser user 6
		{core.NewUserAccount(15919019823465497, "User 7", "user7@foo.com", "", "", "FBID7", "FBTOKEN6"), true},  // Valid user with Facebook credentials (but same token as user6)
	}

	for i, test := range tests {
		if err := dao.Insert(test.user); (err == nil) != test.want {
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
		{core.NewUserAccount(15919019823465498, "User 8", "user8@foo.com", "", "", "FBID3", "FBTOKEN8"), false}, // Same Facebook ID as user 3
		{core.NewUserAccount(15919019823465498, "User 8", "user8@foo.com", "", "", "FBID8", "FBTOKEN8"), true},  // If previous test didn't remove e-mail from user_email_credentials, this will fail
	}

	for i, test := range tests {
		if err := dao.Insert(test.user); (err == nil) != test.want {
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
		{core.NewUserAccount(15918606474806281, "User 9", "user9@foo.com", "", "", "FBID9", "FBTOKEN9"), false}, // Same ID as user 1
		{core.NewUserAccount(15919019823465559, "User 9", "user9@foo.com", "", "", "FBID6", "FBTOKEN6"), false}, // Same FBID as user 6
		{core.NewUserAccount(15919019823465559, "User 9", "user9@foo.com", "", "", "FBID9", "FBTOKEN9"), true},  // If previous inserts didn't remove orphaned rows, this will fail
	}

	for i, test := range tests {
		if err := dao.Insert(test.user); (err == nil) != test.want {
			t.Fatal("Failed at test", i, "with error", err)
		}
	}
}

// Test grace period
func TestInsertFacebookCredentials(t *testing.T) {
	dao := &UserDAO{session}

	var tests = []struct {
		fbid, token string
		id          uint64
		want        bool
	}{
		{"FBID10", "FBTOKEN10", 93, true},
		{"FBID10", "FBTOKEN10", 93, false},
		{"FBID10", "FBTOKEN10", 109, false},
		{"FBID11", "FBTOKEN11", 93, true},
	}

	for i, test := range tests {
		if ok, err := dao.insertFacebookCredentials(test.fbid, test.token, test.id); ok != test.want {
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
		{"abcdefg@foo.com", false},
	}

	for i, test := range tests {
		if user_id, _ := dao.GetIDByEmail(test.email); (user_id != 0) != test.want {
			t.Fatal("Failed at test", i, user_id)
		}
	}
}

func TestGetIDByFacebook(t *testing.T) {

	dao := NewUserDAO(session)

	var tests = []struct {
		fbid string
		want bool
	}{
		{"FBID1", false},
		{"FBID2", true},
		{"FBID3", true},
		{"FBID4", false},
		{"FBID5", false},
		{"FBID6", true},
		{"FBID7", true},
		{"FBID8", true},
		{"FBID9", true},
	}

	for i, test := range tests {
		if user_id, _ := dao.GetIDByFacebookID(test.fbid); (user_id != 0) != test.want {
			t.Fatal("Failed at test", i, user_id)
		}
	}
}

func TestCheckEmailCredentials(t *testing.T) {

	dao := NewUserDAO(session)

	var tests = []struct {
		email, password string
		want            uint64
	}{
		{"user1@foo.com", "12345", 15918606474806281}, // Valid user and password
		{"user1@foo.com", "123456", 0},                // User exists but invalid password
		{"user2@foo.com", "12345", 0},                 // User exists but doesn't have e-mail credentials
		{"user3@foo.com", "12345", 15918606642578453}, // Valid user and password
		{"noexist@foo.com", "12345", 0},               // Invalid user
		{"user1@foo.com", "", 0},                      // User exist but invalid password
	}

	for i, test := range tests {
		if user_id, err := dao.CheckEmailCredentials(test.email, test.password); user_id != test.want {
			t.Fatal("Failed at test", i, "with error", err)
		}
	}
}

func TestCheckValidAccount(t *testing.T) {

	dao := &UserDAO{session}

	var tests = []struct {
		user *core.UserAccount
		want bool
	}{
		{core.NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "12345", "", "", ""), false},             // Valid e-mail
		{core.NewUserAccount(15918606642578452, "User 2", "user2@foo.com", "", "", "FBID2", "FBTOKEN2"), true},      // Remove existing user
		{core.NewUserAccount(15918606642578453, "User 3", "user3@foo.com", "12345", "", "FBID3", "FBTOKEN3"), true}, // Remove existing user
		{core.NewUserAccount(15918606642578454, "User 4", "user4@foo.com", "12345", "", "", ""), true},              // Remove existing user
		{core.NewUserAccount(15919019823465485, "User 5", "user5@foo.com", "12345", "", "", ""), true},              // Remove existing user
		{core.NewUserAccount(15919019823465496, "User 6", "user6@foo.com", "", "", "FBID6", "FBTOKEN6"), true},      // Remove existing user
		{core.NewUserAccount(15919019823465497, "User 7", "user7@foo.com", "", "", "FBID7", "FBTOKEN6"), true},      // Remove existing user
		{core.NewUserAccount(15919019823465498, "User 8", "user8@foo.com", "", "", "FBID8", "FBTOKEN8"), true},      // Remove existing user
		{core.NewUserAccount(15919019823465559, "User 9", "user9@foo.com", "", "", "FBID9", "FBTOKEN9"), true},      // Remove existing user
	}

	dao.DeleteEmailCredentials("user1@foo.com")
	dao.DeleteFacebookCredentials("FBID2")

	for i, test := range tests {
		if ok, err := dao.CheckValidAccount(test.user.Id, false); ok != test.want {
			t.Fatal("Failed at test", i, "with error", err)
		}
	}

}

func TestDelete(t *testing.T) {

	dao := NewUserDAO(session)

	var tests = []struct {
		user *core.UserAccount
		want bool
	}{
		{core.NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "12345", "", "", ""), true},              // Remove existing user
		{core.NewUserAccount(15918606642578452, "User 2", "user2@foo.com", "", "", "FBID2", "FBTOKEN2"), true},      // Remove existing user
		{core.NewUserAccount(15918606642578453, "User 3", "user3@foo.com", "12345", "", "FBID3", "FBTOKEN3"), true}, // Remove existing user
		{core.NewUserAccount(15918606642578454, "User 4", "user4@foo.com", "12345", "", "", ""), true},              // Remove existing user
		{core.NewUserAccount(15919019823465485, "User 5", "user5@foo.com", "12345", "", "", ""), true},              // Remove existing user
		{core.NewUserAccount(15919019823465496, "User 6", "user6@foo.com", "", "", "FBID6", "FBTOKEN6"), true},      // Remove existing user
		{core.NewUserAccount(15919019823465497, "User 7", "user7@foo.com", "", "", "FBID7", "FBTOKEN6"), true},      // Remove existing user
		{core.NewUserAccount(15919019823465498, "User 8", "user8@foo.com", "", "", "FBID8", "FBTOKEN8"), true},      // Remove existing user
		{core.NewUserAccount(15919019823465559, "User 9", "user9@foo.com", "", "", "FBID9", "FBTOKEN9"), true},      // Remove existing user
		{core.NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "12345", "", "", ""), true},              // Remove UNexisting user should return true
	}

	for i, test := range tests {
		if err := dao.Delete(test.user); (err == nil) != test.want {
			t.Fatal("Failed at test", i, "with error", err)
		}
	}
}

func TestUUID(t *testing.T) {

	dao := NewUserDAO(session)
	user := core.NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "12345", "", "", "")

	if err := dao.Insert(user); err != nil {
		t.Fatal(err)
	}

	same_user, _ := dao.Load(user.Id)

	if same_user == nil {
		t.FailNow()
	}

	if user.AuthToken.String() != same_user.AuthToken.String() {
		t.Fatal("Auth are different but should be the same")
	}
}
*/
