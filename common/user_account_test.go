package common

import (
	"testing"
)

func TestValidEmail(t *testing.T) {

	var tests = []struct {
		email string
		want  bool
	}{
		{"foo@example.com", true},          // Ok
		{"foo.bar@example.xyz", true},      // Ok
		{"aa@bb.xy", true},                 // Ok
		{"aa@bb.xyz", true},                // Ok
		{"em", false},                      // Invalid format
		{"em@", false},                     // Invalid format
		{"@", false},                       // Invalid format
		{"@.com", false},                   // Invalid format
		{"em@com", false},                  // Invalid format. Missing point.
		{"e@.com", false},                  // Missing domain part
		{"a@b.x", false},                   // Local part 2 chars minimum
		{"aa@b.x", false},                  // Domain part 2 chars minimum
		{"aa@bb.x", false},                 // Root part 2 chars minimum
		{"aa@bb.wxyz", false},              // More than 3 chars in root part
		{"foo bar@example.com.xyz", false}, // Address has a whitespace
	}

	for i, test := range tests {
		if valid := IsValidEmail(test.email); valid != test.want {
			t.Fatalf("Failed at test %v (email=%v, wanted=%v, obtained=%v)", i, test.email, test.want, valid)
		}
	}

}

// This test is the same as user_dao_test.go TestInsert1
/*func TestValidUserAccounts(t *testing.T) {

	var tests = []struct {
		user *UserAccount
		want bool
	}{
		{NewUserAccount(0, "User 1", "user1@foo.com", "", "", "FBID0", "FBTOKEN0"), false},                     // Invalid ID
		{NewUserAccount(15918606474806281, "", "user1@foo.com", "", "", "FBID0", "FBTOKEN0"), false},           // Invalid name
		{NewUserAccount(15918606474806281, "em", "user1@foo.com", "", "", "FBID0", "FBTOKEN0"), false},         // Invalid name (less than 3 chars)
		{NewUserAccount(15918606474806281, "User 1", "", "", "", "FBID0", "FBTOKEN0"), false},                  // Invalid e-mail
		{NewUserAccount(15918606474806281, "User 1", "em@ail", "", "", "FBID0", "FBTOKEN0"), false},            // Invalid e-mail (no match format)
		{NewUserAccount(15918606474806281, "User 1", "em da@ail.com", "", "", "FBID0", "FBTOKEN0"), false},     // Invalid e-mail (no match format)
		{NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "", "", "", ""), false},                  // There isn't credentials
		{NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "123", "", "", ""), false},               // Invalid password (less than 5 chars)
		{NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "12345", "", "", ""), true},              // Valid user with e-mail credentials
		{NewUserAccount(15918606642578452, "User 2", "user2@foo.com", "", "", "FBID2", ""), false},             // There is facebook id but not token
		{NewUserAccount(15918606642578452, "User 2", "user2@foo.com", "", "", "FBID2", "FBTOKEN2"), true},      // Valid user with Facebook credentials
		{NewUserAccount(15918606642578453, "User 3", "user3@foo.com", "12345", "", "FBID3", "FBTOKEN3"), true}, // Valid user with both credentials
	}

	for i, test := range tests {
		if valid, err := test.user.IsValid(); valid != test.want {
			t.Fatalf("Failed at test %v (wanted=%v, obtained=%v): Error %v", i, test.want, valid, err)
		}
	}
}*/
