package gateway

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func authLDAP(w http.ResponseWriter, r *http.Request) {

	// 改成支持json
	body, _ := ioutil.ReadAll(r.Body)
	var Login struct {
		Username string
		Password string
	}
	code, msg, token := 0, "成功", ""
	if err := json.Unmarshal(body, &Login); err != nil {
		code, msg = 10007, "登录方式错误"
	}

	if Login.Username != "" && Login.Password != "" {
		if err := GS.LdapAuth(Login.Username, Login.Password); err != nil {
			code, msg = 10001, "用户名或者密码错误"
		} else {
			// 使用jwt-token
			if signedToken, err := createToken(&Claims{Username: Login.Username, Password: Login.Password}); err != nil {
				code, msg = 10003, err.Error()
			} else {
				token = signedToken
			}
		}
	} else {
		code, msg = 10002, "用户名或者密码不能为空"
	}
	if data, err := json.Marshal(TokenResult{Code: code, Message: msg, Token: token}); err != nil {
		w.WriteHeader(500)
		log.Panicln("生成json结构体报错。", err.Error())
		w.Write([]byte(""))
	} else {
		w.Write(data)
	}
}

func setPrivateToken(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return
	}
	code, msg := 0, "成功"
	if app, token := r.FormValue("app"), r.FormValue("token"); app != "" && token != "" {
		if err := setAPPToken(r.Header.Get("Authorization-User"), app, token, 0); err != nil {
			code, msg = 10006, "写入redis失败"
		}
	} else {
		code, msg = 10005, "缺少必要的参数"
	}
	if data, err := json.Marshal(JsonResult{Code: code, Message: msg}); err != nil {
		w.WriteHeader(500)
		log.Panicln("生成json结构体报错。", err.Error())
		w.Write([]byte(""))
	} else {
		w.Write(data)
	}
}

func init() {
	http.HandleFunc("/auth/ldap", authLDAP)
	http.HandleFunc("/user/private_token", GS.middleware(setPrivateToken))
}
