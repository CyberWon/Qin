package gateway

import (
	"Qin/pkg/req"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	url2 "net/url"
	"strings"
	"sync"
)

var (
	Director CustomDirector
)

type CustomDirector struct {
}

func beforeRemoveHeader(r *http.Request) {
	// 避免携带的信息被后端服务判断为非法请求
	r.Header.Del("Cookie")
	r.Header.Del("Origin")
	r.Header.Del("Referer")
	r.Header.Del("Sec-Fetch-Mode")
	r.Header.Del("Sec-Fetch-Site")
	r.Header.Del("Sec-Fetch-User")
	r.Header.Del("Sec-fetch-dest")
	r.Header.Del("Sec-Ch-Ua")
	r.Header.Del("Sec-Ch-Ua-Mobile")
	r.Header.Del("Sec-Ch-Ua-Platform")
	r.Header.Del("X-Forwarded-For")
	r.Header.Del("User-Agent")
}

func afterRemoveHeader(r *http.Request) {
	// 反向代理删除header
	r.Header.Del("Authorization-User")
	r.Header.Del("Tenant")
	r.Header.Del("Authorization-URL")
	r.Header.Del("Authorization-Ext")
	r.Header.Del("Authorization-Domain")
}

// None 无认证的情况下，直接去请求
func (director CustomDirector) None(r *http.Request) {

}

func (director CustomDirector) Basic(r *http.Request) {
	if claims, err := vaildateToken(r.Header.Get("Authorization")); err != nil {
		log.Println("验证token失败")
	} else {
		r.SetBasicAuth(claims.Username, claims.Password)
	}
}

func (director CustomDirector) PrivateToken(r *http.Request) {
	user, app := r.Header.Get("Authorization-User"), r.Header.Get("Tenant")
	if token, err := getAPPToken(user, app); err != nil {
		log.Println("没有设置", user, app, "的private token，请先通过接口设置")
	} else {
		r.Header.Set("Authorization", token)
	}
}

func (director CustomDirector) Rancher(r *http.Request) {
	user, app := r.Header.Get("Authorization-User"), r.Header.Get("Tenant")
	if token, err := getAPPToken(user, app); err != nil {
		log.Println("没有设置", user, app, "的private token，请先通过接口设置")
	} else {
		r.Header.Set("Authorization", token)
		r.AddCookie(&http.Cookie{Name: "R_SESS", Value: strings.Replace(token, "Bearer ", "", -1)})
	}
}

// Token 基于token认证的方式
func (director CustomDirector) Token(r *http.Request) {
	user, app, url := r.Header.Get("Authorization-User"), r.Header.Get("Tenant"), r.Header.Get("Authorization-URL")
	ext := strings.Split(r.Header.Get("Authorization-Ext"), " ")
	usernameField, passwordField, tokenTypeField, tokenField := ext[0], ext[1], ext[2], ext[3]
	if token, err := getAPPToken(user, app); err != nil {
		username, password := "", ""
		if claims, err := vaildateToken(r.Header.Get("Authorization")); err != nil {
			log.Error("验证token失败")
		} else {
			username = claims.Username
			password = claims.Password
			data, _ := json.Marshal(map[string]string{usernameField: username, passwordField: password})
			h := req.HTTP{}
			if httpResult := h.Post(url, data); httpResult == nil {
				log.Error("获取token失败")
			} else {
				if accessToken := httpResult[tokenField].(string); accessToken != "" {
					setAPPToken(username, app, tokenTypeField+" "+accessToken, 7200)
					r.Header.Set("Authorization", tokenTypeField+" "+accessToken)
				}
			}
		}
	} else {
		r.Header.Set("Authorization", token)
	}

}

// BKPaaS 蓝鲸运维平台认证
func (director CustomDirector) BKPaaS(r *http.Request) {
	user, url, domain := r.Header.Get("Authorization-User"), r.Header.Get("Authorization-URL"), r.Header.Get("Authorization-Domain")
	token, user := url, user+domain
	r.URL.RawQuery += fmt.Sprintf("&bk_app_code=%s&bk_app_secret=%s&bk_username=%s",
		"gateway", token, user)
	r.Header.Del("Authorization-URL")
}

// BK 蓝鲸cookie认证
func (director CustomDirector) BK(r *http.Request) {
	user, app, url, domain := r.Header.Get("Authorization-User"), "bk", r.Header.Get("Authorization-URL"), r.Header.Get("Authorization-Ext")
	// todo: 这里会导致无法多节点部署。后续解决，做成分布式。
	if _, ok := GS.CookieManager[user]; !ok {
		GS.CookieManager[user] = map[string]req.HTTPCookie{}
	}
	if _, ok := GS.CookieManager[user][app]; !ok {
		gCookieJar, _ := cookiejar.New(nil)
		GS.CookieManager[user][app] = req.HTTPCookie{CookiesJar: gCookieJar, Client: &http.Client{Jar: gCookieJar}}
	}
	c := GS.CookieManager[user][app]
	// 判断是否认证通过
	validate := true
	c.Url(r.URL.String()).Do()
	// 直接请求cmdb的接口并不会跳转，只会返回401错误，这里重新对去访问一下登录页面获取一下csrftoken
	if c.Response.Request != nil && c.Response.Request.Response != nil {
		log.Debug(c.Response.Request.Response.StatusCode, c.Response.Request.Response.Status)
		validate = false
	} else if c.Response.StatusCode == 401 { // 登录失效会进行302跳转，如果没失效，Response.Request.Response的值为nil
		validate = false
		c.Url(url).Do()
	}
	if !validate {
		var m sync.Mutex
		m.Lock()
		// 获取csrftoken
		csrftoken := c.GetCookies("bklogin_csrftoken").Value
		// 模拟登录
		password := ""
		if claims, err := vaildateToken(r.Header.Get("Authorization")); err != nil {
			log.Error("验证token失败")
		} else {
			password = claims.Password
			postData := url2.Values{"csrfmiddlewaretoken": []string{csrftoken},
				"username": []string{fmt.Sprintf("%s%s", user, domain)},
				"password": []string{password}, "next": []string{""}, "app_id": []string{""}}
			c.Url(url).Post(strings.NewReader(postData.Encode()), "urlencoded").Do()
			log.Debug("登录蓝鲸", c.Response.StatusCode, c.Response.Status)
			c.Url(r.URL.String()).Do().SetCookie(r)
		}
		m.Unlock()
	} else {
		c.SetCookie(r)
	}
	// 蓝鲸SaaS会获取header里面的X-CSRFToken，不然无法POST，这个值在cookie里面
	if ok := strings.Contains(r.URL.String(), "itsm"); ok {
		CSRFToken := c.GetCookies("bkitsm_csrftoken")
		if CSRFToken != nil {
			r.Header.Set("X-CSRFToken", CSRFToken.Value)
		}
	}
}
