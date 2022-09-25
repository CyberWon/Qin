package gateway

import (
	"Qin/pkg/req"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	url2 "net/url"
	"strings"
)

var (
	Director CustomDirector
)

type CustomDirector struct {
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

func (director CustomDirector) Token(r *http.Request) {
	//log.Println(r.Header.Get("Authorization"))
}

// Archery 数据库审计平台
func (director CustomDirector) Archery(r *http.Request) {

	user, app, url := r.Header.Get("Authorization-User"), r.Header.Get("Tenant"), r.Header.Get("Authorization-URL")
	if token, err := getAPPToken(user, app); err != nil {
		username, password := "", ""
		if claims, err := vaildateToken(r.Header.Get("Authorization")); err != nil {
			log.Println("验证token失败")
		} else {
			username = claims.Username
			password = claims.Password
			data, _ := json.Marshal(map[string]string{"username": username, "password": password})
			h := req.HTTP{}
			if httpResult := h.Post(url, data); httpResult == nil {
				log.Panic("获取token失败")
			} else {
				if accessToken := httpResult["access"].(string); accessToken != "" {
					setAPPToken(username, app, "Bearer "+accessToken, 7200)
					r.Header.Set("Authorization", "Bearer "+accessToken)
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
	user, app, url, domain := r.Header.Get("Authorization-User"), r.Header.Get("Tenant"), r.Header.Get("Authorization-URL"), r.Header.Get("Authorization-Domain")
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
	// 登录失效会进行302跳转，如果没失效，Response.Request.Response的值为nil
	if c.Response.Request != nil && c.Response.Request.Response != nil {
		validate = false
	}
	if !validate {
		// 获取csrftoken
		csrftoken := c.GetCookies("bklogin_csrftoken").Value
		// 模拟登录
		password := ""
		if claims, err := vaildateToken(r.Header.Get("Authorization")); err != nil {
			log.Println("验证token失败")
		} else {
			password = claims.Password
			postData := url2.Values{"csrfmiddlewaretoken": []string{csrftoken},
				"username": []string{fmt.Sprintf("%s%s", user, domain)},
				"password": []string{password}, "next": []string{""}, "app_id": []string{""}}
			c.Url(url).Post(strings.NewReader(postData.Encode()), "urlencoded").Do()
			c.Url(r.URL.String()).Do().SetCookie(r)
		}
	} else {
		c.SetCookie(r)
	}
	// 蓝鲸SaaS会获取header里面的X-CSRFToken，不然无法POST，这个值在cookie里面
	if ok := strings.Contains(r.URL.String(), "itsm"); ok {
		r.Header.Set("X-CSRFToken", c.GetCookies("bkitsm_csrftoken").Value)
	}
}
