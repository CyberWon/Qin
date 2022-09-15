package gateway

import (
	"MicroOps/pkg/req"
	"crypto/tls"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reflect"
	"strings"
)

var (
	GS Gateway
)

func init() {

	// 读取相应配置文件读取
	if f, err := os.OpenFile("config.yml", os.O_RDONLY, 0444); err != nil {
		log.Panic("打开配置文件失败")
	} else {
		if data, err := io.ReadAll(f); err != nil {
			log.Panic("读取文件失败")
		} else {
			if err := yaml.Unmarshal(data, &GS.Config); err != nil {
				log.Panic("格式化配置文件失败")
			}
		}
	}

	// 日志格式，为了符合云原生，就不写入到文件了。
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	if conn, err := redis.Dial("tcp", GS.Config.Redis.Addr); err != nil {
		log.Panic("连接redis失败，请检查redis服务或者网络状态")
	} else {
		GS.Cache = conn
		if GS.Config.Redis.Auth != "" {
			if _, err := GS.Cache.Do("auth", GS.Config.Redis.Auth); err != nil {
				log.Panic("redis认证错误，请检查密码是否正常")
			}
		}
	}

}

type Handle func(http.ResponseWriter, *http.Request)

func (g *Gateway) loadApp() {
	for _, config := range g.Config.Proxy {
		log.Println("加载", config.App, "配置")
		// 新建反向代理服务
		if rp, err := NewReverseGateway(config); err != nil {
			log.Panic(err)
		} else {
			http.HandleFunc("/"+config.App+"/", g.middleware(rp.ServeHTTP))
		}
	}
}

func (g *Gateway) Start() error {

	g.CookieManager = make(map[string]map[string]req.HTTPCookie)

	// 加载第三方应用
	g.loadApp()

	// 链接到数据源
	g.ldapConn()

	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", GS.Config.Server.Port), nil)

	// 释放一些资源
	defer g.LDAP.Conn.Close()

	return err
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	// 找到第三个/字符下标
	bIndex := strings.Index(b[1:], "/") + 1
	switch {
	case aslash && bslash:
		return a + b[bIndex:]
	case !aslash && !bslash:
		return a + "/" + b[bIndex:]
	}
	return a + b[bIndex:]
}

func RewriteURL(a, b *url.URL) (path, rawpath, uri string) {
	if a.RawPath == "" && b.RawPath == "" {
		s := singleJoiningSlash(a.Path, b.Path)
		return s, s, s
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	p := singleJoiningSlash(apath, bpath)
	return p, p, p
}

func NewReverseGateway(proxy Proxy) (*httputil.ReverseProxy, error) {
	if target, err := url.Parse(proxy.URL); err != nil {
		return nil, err
	} else {
		targetQuery := target.RawQuery
		// 有一些反向代理的网站ssl证书会有问题，这里可以自定义是否检测
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: proxy.Insecure},
		}

		// request的一些构造。
		director := func(req *http.Request) {
			req.Host = target.Host
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path, req.URL.RawPath, req.RequestURI = RewriteURL(target, req.URL)
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
			log.Println(req.URL)
			// 添加访问的app信息到header头部
			req.Header.Set("Authorization-App", proxy.App)
			req.Header.Set("Authorization-URL", proxy.AuthorizationURL)
			req.Header.Set("Authorization-Domain", proxy.AuthorizationDomain)
			// 利用反射去调用自定义的Director,来修改对应的headers，认证信息等。
			customDirector := reflect.ValueOf(&Director)
			method := customDirector.MethodByName(proxy.Authorization)
			params := []reflect.Value{
				reflect.ValueOf(req), // 方法参数
			}
			method.Call(params) // 返回的是字符串
		}

		return &httputil.ReverseProxy{Director: director, Transport: transport, ModifyResponse: func(response *http.Response) error {

			return nil
		}}, nil
	}
}