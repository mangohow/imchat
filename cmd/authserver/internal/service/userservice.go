package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mangohow/imchat/cmd/authserver/internal/dao"
	"github.com/mangohow/imchat/cmd/authserver/internal/log"
	"github.com/mangohow/imchat/cmd/authserver/internal/rdsconn"
	"github.com/mangohow/imchat/pkg/consts/redisconsts"
	"github.com/mangohow/imchat/pkg/model"
	"github.com/mangohow/imchat/pkg/utils"
	"github.com/sirupsen/logrus"
)

type UserService struct {
	userDao *dao.UserDao
	redis *redis.Client
	logger *logrus.Logger
}

func NewUserService() *UserService {
	return &UserService{
		userDao: dao.NewUserDao(),
		redis: rdsconn.RedisConn(),
		logger: log.Logger(),
	}
}

var(
	CodeInvalidError = errors.New("验证码有误")

	CodeGetFrequentError = errors.New("验证码获取频繁")

	PhoneHadBeenRegisteredError = errors.New("该手机已经被注册")
	UsernameUnavailable = errors.New("用户名重复")
)

// GetPhoneVerificationCode 获取手机验证码
// 为了保证服务不会被频繁调用，在一分钟内同一个电话号码只能获取一次
func (s *UserService) GetPhoneVerificationCode(phone string) (string, error) {
	// 先检查在一分钟内是否已经发送了验证码
	codeKey := redisconsts.LoginPhoneKey + phone
	left, _ := s.redis.TTL(context.Background(), codeKey).Result()
	s.logger.Debugf("leftTime:%d", left)
	// redis中存在验证码, 查看剩余时间
	// 如果key不存在，则left == -2，因此下面结果不成立，即会发送验证码
	if left > redisconsts.PhoneCodeDuration - redisconsts.PhoneCodeResendDuration  {
		// 不能重复发送验证码
		return "", CodeGetFrequentError
	}

	// 生成验证码，并保存到redis
	code := utils.GetRandomCode(6)
	if cmd := s.redis.SetEX(context.Background(), codeKey, code, redisconsts.PhoneCodeDuration); cmd.Err() != nil {
		return "", cmd.Err()
	}

	return code, nil
}

func (s *UserService) CreateUser(user *model.UserRegister) error {
	// 先查询该手机号是否已经注册过了
	if ok := s.userDao.CheckUserRegistered(user.Phone); ok {
		return PhoneHadBeenRegisteredError
	}
	// 查询用户名是否可用
	if ok := s.userDao.CheckUsernameAvailable(user.Username); ok {
		return UsernameUnavailable
	}

	// 从redis查询code
	key := redisconsts.LoginPhoneKey + user.Phone
	code, err := s.redis.Get(context.Background(), key).Result()
	if err != nil || code != user.Code {
		return CodeInvalidError
	}

	id, err := s.generateUserId()
	if err != nil {
		s.logger.Errorf("generate user id error:%v", err)
		return err
	}
	
	now := time.Now()
	usr := &model.User{
		Id: id,
		Username: user.Username,
		Password: user.Password,
		CreateTime: now,
		RemoveTime: now,
		ReleaseTime: now,
		BanedTime: now,
		PhoneNum: user.Phone,
	}
	userinfo := &model.Userinfo{
		Id:        usr.Id,
		Username:  usr.Username,
		Nickname:  utils.GetRandomAlphString(10),
		Avatar:    "",
		Signature: "",
		Gender:    0,
		Birthday:  now,
		Address:   "",
		HomeAddr:  "",
		PhoneNum:  user.Phone,
		Email:     "",
	}

	return s.userDao.CreateUser(usr, userinfo)
}

// 生成用户ID
// 用户ID由 时间戳(48bit) + 计数器(16bit) 生成 ，借助于redis的自增
func (s *UserService) generateUserId() (int64, error) {
	n, err := s.redis.Incr(context.Background(), redisconsts.UserCounterKey).Result()
	if err != nil {
		return 0, err
	}
	if n >= math.MaxUint16 {
		s.redis.Set(context.Background(), redisconsts.UserCounterKey, 0, 0)
	}

	id := time.Now().Unix()

	id <<= 16
	id |= int64(uint16(n))

	return id, nil
}

func (s *UserService) CheckUserLogin(login *model.UserLogin) (user *model.User, userinfo *model.Userinfo, token string, err error) {
	user, userinfo, err = s.userDao.GetUserAndInfoByUsernameAndPassword(login)
	if err != nil {
		return
	}

	err = s.userAvailableCheck(user)
	if err != nil {
		return
	}

	// 生成token
	token, err = utils.CreateToken(user.Id, user.Username)
	if err != nil {
		return
	}

	idStr := strconv.Itoa(int(user.Id))

	userData, _ := json.Marshal(user)
	userInfoData, _ := json.Marshal(userinfo)
	pipeline := s.redis.Pipeline()

	// 将token保存到redis, 过期时间30s
	if err = pipeline.Set(context.Background(), redisconsts.TokenKey + idStr, token, redisconsts.TokenExpireDuration).Err(); err != nil {
		s.logger.Errorf("save token error:%v", err)
		return
	}
	// 将数据缓存到redis中，过期时间30m


	pipeline.Set(context.Background(), redisconsts.UserKey + idStr, userData, redisconsts.UserCacheExpireDuration)
	pipeline.Set(context.Background(), redisconsts.UserInfoKey + idStr, userInfoData, redisconsts.UserCacheExpireDuration)
	if _, err = pipeline.Exec(context.Background()); err != nil {
		s.logger.Errorf("cache data error:%v", err)
	}

	return
}


type userStatus int

const (
	UserStatusNormal    = iota // 账号正常
	UserStatusBanned           // 账号封禁
	UserStatusDestroyed        // 账号销毁
)

var (
	UserBanned = errors.New("user banned")
	UserDestroy = errors.New("user destroyed")
)

func (s *UserService) userAvailableCheck(user *model.User) error {
	status := userStatus(user.Status)
	switch status {
	case UserStatusBanned:
		if canBeReleasing(user.ReleaseTime) {
			return s.userDao.UpdateUserStatus(user.Id, 0)
		} else {
			return UserBanned
		}
	case UserStatusDestroyed:
		return UserDestroy
	}

	return nil
}

func (s *UserService) GetUserInfo(id int64) *model.Userinfo {
	// 先从缓存查询
	ret := s.redis.Get(context.Background(), redisconsts.UserInfoKey+strconv.Itoa(int(id)))
	// 查询到
	bytes, err := ret.Bytes()
	if err != nil && err != redis.Nil {
		userinfo := new(model.Userinfo)
		json.Unmarshal(bytes, userinfo)
		return userinfo
	}

	// 查询不到再查询数据库
	return s.userDao.GetUserInfo(id)
}

func canBeReleasing(releaseTime time.Time) bool {
	if releaseTime.Unix() <= time.Now().Unix() {
		return true
	}

	return false
}


