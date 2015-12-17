package common

func CreateFakeUsers(dao UserDAO) {

	user1 := NewUserAccount(15918606474806282, "User 1", "user1@foo.com", "12345", "", "", "")
	user2 := NewUserAccount(15918606642578453, "User 2", "user2@foo.com", "12345", "", "", "")
	user3 := NewUserAccount(15918606642578454, "User 3", "user3@foo.com", "12345", "", "", "")
	user4 := NewUserAccount(15918606642578455, "User 4", "user4@foo.com", "12345", "", "", "")
	user5 := NewUserAccount(15919019823465486, "User 5", "user5@foo.com", "12345", "", "", "")
	user6 := NewUserAccount(15919019823465491, "User 6", "user6@foo.com", "12345", "", "", "")
	user7 := NewUserAccount(15919019823465497, "User 7", "user7@foo.com", "12345", "", "", "")
	user8 := NewUserAccount(15919019823465498, "User 8", "user8@foo.com", "12345", "", "", "")

	dao.Insert(user1)
	dao.Insert(user2)
	dao.Insert(user3)
	dao.Insert(user4)
	dao.Insert(user5)
	dao.Insert(user6)
	dao.Insert(user7)
	dao.Insert(user8)

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
