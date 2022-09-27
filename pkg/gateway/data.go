package gateway

import (
	"Qin/pkg/req"
	"github.com/go-ldap/ldap/v3"
	"github.com/gomodule/redigo/redis"
)

type Gateway struct {
	Config        Config
	LDAP          LDAP
	Cache         redis.Conn
	CookieManager map[string]map[string]req.HTTPCookie
}

type LDAP struct {
	Conn *ldap.Conn
}

// Config 配置文件结构体
type Config struct {
	Tenants []Tenant
	Ldap    struct {
		Host       string `yaml:"host"`
		Port       int    `yaml:"port"`
		BindDn     string `yaml:"bind_dn"`
		UserBaseDN string `yaml:"user_base_dn"`
		Password   string `yaml:"password"`
	} `yaml:"ldap"`
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Auth struct {
		Secret string `yaml:"secret"`
		Expire int    `yaml:"expire"`
	} `yaml:"auth"`
	Redis struct {
		Addr string
		DB   int
		Auth string
	}
	Default struct {
		Tenant string
	}
}
type Tenant struct {
	Tenant              string
	URL                 string
	Authorization       string
	AuthorizationURL    string `yaml:"authorization_url"`
	AuthorizationDomain string `yaml:"authorization_domain"`
	Insecure            bool
}

// JsonResult 定义一下接口返回的结构体
type JsonResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// TokenResult Token返回的结构体
type TokenResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Token   string `json:"token"`
}

type UserResult struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    UserStruct `json:"data"`
}
