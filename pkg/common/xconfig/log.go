package xconfig

type LogConfig struct {
	Level       string `yaml:"level"`
	FilePath    string `yaml:"filePath"`
	FileName    string `yaml:"fileName"`
	MaxFileSize uint64 `yaml:"maxFileSize"`
	ToFile      bool   `yaml:"toFile"`
	Formatter   string `yaml:"formatter"`
	Caller      bool   `yaml:"caller"`
}
