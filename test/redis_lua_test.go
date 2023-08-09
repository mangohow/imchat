package test

import (
	"context"
	"testing"

	"github.com/mangohow/imchat/pkg/common/xconfig"
	"github.com/mangohow/imchat/pkg/common/xredis"
	"github.com/mangohow/imchat/pkg/consts/redisconsts"
)

func TestRedisLua(t *testing.T) {
	redis, err := xredis.NewRedisInstance(&xconfig.RedisConfig{
		Addr:         "192.168.44.100:6379",
		PoolSize:     10,
		MinIdleConns: 5,
		Password:     "",
		DB:           0,
	})

	if err != nil {
		t.Fatal(err)
	}

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

	key := redisconsts.FriendsKey + "1687054235708952576"
	res := redis.Eval(context.Background(), script, []string{key, redisconsts.ChatServerClientKey})
	result, err := res.Int64Slice()
	if err != nil {
		t.Errorf("error:%v", err)
	} else {
		t.Log(result)
	}
}
