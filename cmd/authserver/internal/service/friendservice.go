package service

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/elliotchance/pie/v2"
	"github.com/go-redis/redis/v8"
	"github.com/mangohow/imchat/cmd/authserver/internal/dao"
	"github.com/mangohow/imchat/cmd/authserver/internal/log"
	"github.com/mangohow/imchat/cmd/authserver/internal/rdsconn"
	"github.com/mangohow/imchat/pkg/consts/redisconsts"
	"github.com/mangohow/imchat/pkg/model"
	"github.com/mangohow/imchat/pkg/utils"
	"github.com/sirupsen/logrus"
)

type FriendService struct {
	dao *dao.FriendDao
	userDao *dao.UserDao
	redis *redis.Client
	logger *logrus.Logger
}

func NewContactFriendService() *FriendService {
	return &FriendService{
		dao: dao.NewContactFriendDao(),
		userDao: dao.NewUserDao(),
		redis: rdsconn.RedisConn(),
		logger: log.Logger(),
	}
}

func (s *FriendService) GetAllFriends(userId int64) ([]*model.FriendDTO, error) {
	// 1. 从数据库查询该用户的所有朋友
	friends := s.dao.FindFriendsById(userId)
	if len(friends) == 0 {
		return nil, nil
	}

	// 2. 将ids缓存到redis
	friendIds := pie.Map(friends, func(t *model.Friend) int64 {
		return t.FriendId
	})
	friendsKey := redisconsts.FriendsKey+strconv.Itoa(int(userId))
	pip := s.redis.Pipeline()
	pip.SAdd(context.Background(), friendsKey, utils.ToInterfaceSlice(friendIds)...)
	pip.Expire(context.Background(), friendsKey, redisconsts.DefaultCacheDuration)
	_, err := pip.Exec(context.Background())
	if err != nil {
		s.logger.Errorf("save friends error:%v", err)
		return nil, err
	}

	// 3. 获取朋友信息
	keys := pie.Map(friendIds, func(id int64) string {
		return redisconsts.UserInfoKey + strconv.Itoa(int(id))
	})

	// 3.1 先从缓存中查询
	friendInfos := make([]*model.FriendDTO, 0, len(friendIds))
	// 先从缓存中查询
	r := s.redis.MGet(context.Background(), keys...)
	infosStr, _ := r.Result()
	for i := range infosStr {
		if infosStr[i] == nil {
			continue
		}
		friend := new(model.FriendDTO)
		friend.Userinfo = new(model.Userinfo)
		err := json.Unmarshal(utils.String2Bytes(infosStr[i].(string)), &friend.Userinfo)
		if err == nil {
			friendInfos = append(friendInfos, friend)
		}
	}

	for {
		// 3.2 如果从缓存中查找到所有key，直接返回
		if len(friendInfos) == cap(friendInfos) {
			break
		}

		// 3.3 否则没有查到的 从数据库中查询
		found := pie.Map(friendInfos, func(t *model.FriendDTO) int64 {
			return t.Userinfo.Id
		})

		needed, _ := pie.Diff(found, friendIds)

		infos := s.userDao.GetUserInfoInIds(needed)
		for i := range infos {
			friend := new(model.FriendDTO)
			friend.Userinfo = infos[i]
			friendInfos = append(friendInfos, friend)
		}

		// 4. 将userinfo缓存到数据库
		pip := s.redis.Pipeline()
		for _, info := range friendInfos {
			bytes, err := json.Marshal(info.Userinfo)
			if err == nil {
				pip.SetEX(context.Background(), redisconsts.UserInfoKey + strconv.Itoa(int(info.Userinfo.Id)), bytes, redisconsts.DefaultCacheDuration)
			}
		}
		pip.Exec(context.Background())
		break
	}

	m := make(map[int64]*model.Friend)
	for i := range friends {
		m[friends[i].FriendId] = friends[i]
	}

	pie.Each(friendInfos, func(dto *model.FriendDTO) {
		friend := m[dto.Userinfo.Id]
		if friend != nil {
			dto.Remark = friend.Remark
			dto.Auth = int(friend.Auth)
			dto.CreateTime = friend.CreateTime
		}
	})

	return friendInfos, nil
}

func (s *FriendService) GetOnlineFriends(id int64) []int64 {
	script :=
`
local res = redis.call('SMEMBERS', KEYS[1])
if not res then
	return nil
end

local retVal = {}
for i, v in pairs(res) do
	local r = redis.call('EXISTS', KEYS[2]..v)
	if r == 1 then
		table.insert(retVal, v)
	end
end

return retVal
`

	key := redisconsts.FriendsKey + strconv.Itoa(int(id))
	res := s.redis.Eval(context.Background(), script, []string{key, redisconsts.ChatServerClientKey})
	result, err := res.Int64Slice()
	if err != nil {
		s.logger.Errorf("get online friends error:%v", err)
	}

	return result
}



