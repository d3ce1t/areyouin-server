package common

import (
	"strconv"
)

func AddFriendsToFbTestUserOne(dao UserDAO) {

	user2 := NewUserAccount(15918606642578452, "User 2", "user2@foo.com", "12345", "", "FBID2", "FBTOKEN2")
	user3 := NewUserAccount(15918606642578453, "User 3", "user3@foo.com", "12345", "", "FBID3", "FBTOKEN3")
	user4 := NewUserAccount(15918606642578454, "User 4", "user4@foo.com", "12345", "", "", "")
	user5 := NewUserAccount(15919019823465485, "User 5", "user5@foo.com", "12345", "", "", "")
	user6 := NewUserAccount(15919019823465496, "User 6", "user6@foo.com", "12345", "", "FBID6", "FBTOKEN6")
	user7 := NewUserAccount(15919019823465497, "User 7", "user7@foo.com", "12345", "", "FBID7", "FBTOKEN7")
	user8 := NewUserAccount(15919019823465498, "User 8", "user8@foo.com", "12345", "", "FBID8", "FBTOKEN8")

	dao.AddFriend(23829049541395456, user2.AsFriend(), 0)
	dao.AddFriend(23829049541395456, user3.AsFriend(), 0)
	dao.AddFriend(23829049541395456, user4.AsFriend(), 0)
	dao.AddFriend(23829049541395456, user5.AsFriend(), 0)
	dao.AddFriend(23829049541395456, user6.AsFriend(), 0)
	dao.AddFriend(23829049541395456, user7.AsFriend(), 0)
	dao.AddFriend(23829049541395456, user8.AsFriend(), 0)
}

func DeleteFakeusers(dao UserDAO) {

	user1 := NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "12345", "", "", "")
	user2 := NewUserAccount(15918606642578452, "User 2", "user2@foo.com", "12345", "", "FBID2", "FBTOKEN2")
	user3 := NewUserAccount(15918606642578453, "User 3", "user3@foo.com", "12345", "", "FBID3", "FBTOKEN3")
	user4 := NewUserAccount(15918606642578454, "User 4", "user4@foo.com", "12345", "", "", "")
	user5 := NewUserAccount(15919019823465485, "User 5", "user5@foo.com", "12345", "", "", "")
	user6 := NewUserAccount(15919019823465496, "User 6", "user6@foo.com", "12345", "", "FBID6", "FBTOKEN6")
	user7 := NewUserAccount(15919019823465497, "User 7", "user7@foo.com", "12345", "", "FBID7", "FBTOKEN7")
	user8 := NewUserAccount(15919019823465498, "User 8", "user8@foo.com", "12345", "", "FBID8", "FBTOKEN8")
	user9 := NewUserAccount(15919019823465559, "User 9", "user9@foo.com", "", "", "FBID9", "FBTOKEN9")

	dao.Delete(user1)
	dao.Delete(user2)
	dao.Delete(user3)
	dao.Delete(user4)
	dao.Delete(user5)
	dao.Delete(user6)
	dao.Delete(user7)
	dao.Delete(user8)
	dao.Delete(user9)
}

func CreateFakeUsers(dao UserDAO) {

	user1 := NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "12345", "", "", "")
	user2 := NewUserAccount(15918606642578452, "User 2", "user2@foo.com", "12345", "", "FBID2", "FBTOKEN2")
	user3 := NewUserAccount(15918606642578453, "User 3", "user3@foo.com", "12345", "", "FBID3", "FBTOKEN3")
	user4 := NewUserAccount(15918606642578454, "User 4", "user4@foo.com", "12345", "", "", "")
	user5 := NewUserAccount(15919019823465485, "User 5", "user5@foo.com", "12345", "", "", "")
	user6 := NewUserAccount(15919019823465496, "User 6", "user6@foo.com", "12345", "", "FBID6", "FBTOKEN6")
	user7 := NewUserAccount(15919019823465497, "User 7", "user7@foo.com", "12345", "", "FBID7", "FBTOKEN7")
	user8 := NewUserAccount(15919019823465498, "User 8", "user8@foo.com", "12345", "", "FBID8", "FBTOKEN8")
	user9 := NewUserAccount(15919019823465559, "User 9", "user9@foo.com", "", "", "FBID9", "FBTOKEN9")

	dao.Insert(user1)
	dao.Insert(user2)
	dao.Insert(user3)
	dao.Insert(user4)
	dao.Insert(user5)
	dao.Insert(user6)
	dao.Insert(user7)
	dao.Insert(user8)
	dao.Insert(user9)

	dao.AddFriend(user1.Id, user2.AsFriend(), 0)
	dao.AddFriend(user1.Id, user3.AsFriend(), 0)
	dao.AddFriend(user1.Id, user4.AsFriend(), 0)
	dao.AddFriend(user1.Id, user5.AsFriend(), 0)
	dao.AddFriend(user1.Id, user6.AsFriend(), 0)
	dao.AddFriend(user1.Id, user7.AsFriend(), 0)
	dao.AddFriend(user1.Id, user8.AsFriend(), 0)

	dao.AddFriend(user2.Id, user1.AsFriend(), 0)
	dao.AddFriend(user2.Id, user3.AsFriend(), 0)
	dao.AddFriend(user2.Id, user4.AsFriend(), 0)

	dao.AddFriend(user3.Id, user1.AsFriend(), 0)
	dao.AddFriend(user3.Id, user2.AsFriend(), 0)
	dao.AddFriend(user3.Id, user4.AsFriend(), 0)

	dao.AddFriend(user4.Id, user1.AsFriend(), 0)
	dao.AddFriend(user4.Id, user2.AsFriend(), 0)
	dao.AddFriend(user4.Id, user3.AsFriend(), 0)

	dao.AddFriend(user5.Id, user1.AsFriend(), 0)

	dao.AddFriend(user6.Id, user1.AsFriend(), 0)

	dao.AddFriend(user7.Id, user1.AsFriend(), 0)

	dao.AddFriend(user8.Id, user1.AsFriend(), 0)
}

func Create100Users(dao UserDAO) {

	idgen := NewIDGen(1)

	user10 := NewUserAccount(idgen.GenerateID(), "User 10", "user10@foo.com", "12345", "", "", "")
	dao.Insert(user10)

	for i := 11; i < 110; i++ {
		user := NewUserAccount(idgen.GenerateID(), "User "+strconv.Itoa(i), "user"+strconv.Itoa(i)+"@foo.com", "12345", "", "", "")
		dao.Insert(user)
		dao.MakeFriends(user10.AsFriend(), user.AsFriend())
	}

	//dao.AddFriend(user1.Id, user2.AsFriend(), 0)
}

func Delete100Users(dao UserDAO) {
	for i := 10; i < 110; i++ {
		user, _ := dao.LoadByEmail("user" + strconv.Itoa(i) + "@foo.com")
		dao.Delete(user)
	}
}
