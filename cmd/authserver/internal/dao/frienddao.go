package dao

import "github.com/mangohow/imchat/pkg/model"

// 联系人dao，获取用户好友信息

type FriendDao struct {
}

func NewContactFriendDao() *FriendDao {
	return &FriendDao{
	}
}

func (c *FriendDao) FindFriendInfosByUserId(userId int64) (infos []model.Friend, err error) {

	return
}

// FindFriendsById 从数据库中查询所有联系人
func (c *FriendDao) FindFriendsById(id int64) (friends []*model.Friend) {
	mysqlDB.Table("t_friend").Where("user_id = ?", id).Find(&friends)
	return
}