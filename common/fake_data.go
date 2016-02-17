package common

import (
//"strconv"
)

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

func CreateFakeUsers(userDAO UserDAO, friendDAO FriendDAO) {

	user1 := NewUserAccount(15918606474806281, "User 1", "user1@foo.com", "12345", "", "", "")
	user2 := NewUserAccount(15918606642578452, "User 2", "user2@foo.com", "12345", "", "FBID2", "FBTOKEN2")
	user3 := NewUserAccount(15918606642578453, "User 3", "user3@foo.com", "12345", "", "FBID3", "FBTOKEN3")
	user4 := NewUserAccount(15918606642578454, "User 4", "user4@foo.com", "12345", "", "", "")
	user5 := NewUserAccount(15919019823465485, "User 5", "user5@foo.com", "12345", "", "", "")
	user6 := NewUserAccount(15919019823465496, "User 6", "user6@foo.com", "12345", "", "FBID6", "FBTOKEN6")
	user7 := NewUserAccount(15919019823465497, "User 7", "user7@foo.com", "12345", "", "FBID7", "FBTOKEN7")
	user8 := NewUserAccount(15919019823465498, "User 8", "user8@foo.com", "12345", "", "FBID8", "FBTOKEN8")
	user9 := NewUserAccount(15919019823465559, "User 9", "user9@foo.com", "", "", "FBID9", "FBTOKEN9")

	userDAO.Insert(user1)
	userDAO.Insert(user2)
	userDAO.Insert(user3)
	userDAO.Insert(user4)
	userDAO.Insert(user5)
	userDAO.Insert(user6)
	userDAO.Insert(user7)
	userDAO.Insert(user8)
	userDAO.Insert(user9)

	friendDAO.MakeFriends(user1, user2)
	friendDAO.MakeFriends(user1, user3)
	friendDAO.MakeFriends(user1, user4)
	friendDAO.MakeFriends(user1, user5)
	friendDAO.MakeFriends(user1, user6)
	friendDAO.MakeFriends(user1, user7)
	friendDAO.MakeFriends(user1, user8)

	friendDAO.MakeFriends(user2, user3)
	friendDAO.MakeFriends(user2, user4)

	friendDAO.MakeFriends(user3, user4)
}

/*func Create100Users(dao UserDAO) {

	idgen := NewIDGen(1)

	user10 := NewUserAccount(idgen.GenerateID(), "User 10", "user10@foo.com", "12345", "", "", "")
	dao.Insert(user10)

	for i := 11; i < 110; i++ {
		user := NewUserAccount(idgen.GenerateID(), "User "+strconv.Itoa(i), "user"+strconv.Itoa(i)+"@foo.com", "12345", "", "", "")
		dao.Insert(user)
		dao.MakeFriends(user10, user)
	}

	//dao.AddFriend(user1.Id, user2.AsFriend(), 0)
}*/

/*func Delete100Users(dao UserDAO) {
	for i := 10; i < 110; i++ {
		user, _ := dao.LoadByEmail("user" + strconv.Itoa(i) + "@foo.com")
		dao.Delete(user)
	}
}*/
