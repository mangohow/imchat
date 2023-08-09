package xconfig

type MysqlConfig struct {
	DataSourceName string `yaml:"dataSourceName"`
	MaxOpenConns   int    `yaml:"maxOpenConns"`
	MaxIdleConns   int    `yaml:"maxIdleConns"`
}
