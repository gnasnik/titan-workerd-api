package config

var Cfg Config

type Config struct {
	Mode          string
	ApiListen     string
	DatabaseURL   string
	SecretKey     string
	EtcdAddresses []string
	EtcdUser      string
	EtcdPassword  string
	IpDataCloud   IpDataCloudConfig
}

type IpDataCloudConfig struct {
	Url string
	Key string
}
