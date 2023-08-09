package xconfig

type RedisConfig struct {
	Addr         string
	PoolSize     uint32
	MinIdleConns uint32
	Password     string
	DB           uint32
}
