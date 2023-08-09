package dao

import (
	"github.com/mangohow/imchat/cmd/authserver/internal/log"
	"github.com/mangohow/imchat/pkg/model"
	"github.com/sirupsen/logrus"
)

/*
* @Author: mgh
* @Date: 2022/2/24 12:39
* @Desc:
 */

type UserDao struct {
	logger *logrus.Logger
}

func NewUserDao() *UserDao {
	return &UserDao{
		logger: log.Logger(),
	}
}

func (d *UserDao) CreateUser(user *model.User, userinfo *model.Userinfo) (err error) {
	db := mysqlDB.Begin()
	if err = db.Table("t_user").Create(user).Error; err != nil {
		db.Rollback()
		return err
	}
	d.logger.Debug("id:", user.Id)
	userinfo.Id = user.Id
	if err = db.Table("t_userinfo").Create(userinfo).Error; err != nil {
		db.Rollback()
		return err
	}
	db.Commit()
	return nil
}

func (d *UserDao) GetUserAndInfoByUsernameAndPassword(login *model.UserLogin) (*model.User, *model.Userinfo, error) {
	user := new(model.User)
	userinfo := new(model.Userinfo)

	err := mysqlDB.Table("t_user").Where("username = ? and password = ?", login.Username, login.Password).First(user).Error
	if err != nil {
		return nil, nil, err
	}

	err = mysqlDB.Table("t_userinfo").Where("id = ?", user.Id).First(userinfo).Error

	return user, userinfo, err
}

func (d *UserDao) UpdateUserStatus(id int64, status int) error {
	return mysqlDB.Table("t_user").Where("id = ?", id).Update("status", status).Error
}

func (d *UserDao) GetUserInfo(id int64) *model.Userinfo {
	userinfo := new(model.Userinfo)
	mysqlDB.Table("t_userinfo").Where("id = ?", id).First(userinfo)
	return userinfo
}

func (d *UserDao) GetUserInfoInIds(needed []int64) (infos []*model.Userinfo) {
	mysqlDB.Table("t_userinfo").Where("id IN ?", needed).Find(&infos)
	return
}

func (d *UserDao) CheckUserRegistered(phone string) bool {
	count := 0
	mysqlDB.Raw(`SELECT COUNT(*) FROM t_user WHERE phone_num = ?;`, phone).First(&count)
	return count != 0
}

func (d *UserDao) CheckUsernameAvailable(username string) bool {
	count := 0
	mysqlDB.Raw(`SELECT COUNT(*) FROM t_user WHERE username = ?;`, username).First(&count)
	return count != 0
}
