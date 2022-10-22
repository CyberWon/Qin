package gateway

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type UserStruct struct {
	User string `json:"user"`
	Mail string `json:"mail"`
	Name string `json:"name"`
}

// HTTPApiResult 公用返回方法
func HTTPApiResult(w http.ResponseWriter, result interface{}) {
	if data, err := json.Marshal(result); err != nil {
		w.WriteHeader(500)
		log.Panicln("生成json结构体报错。", err.Error())
		w.Write([]byte(""))
	} else {
		w.Write(data)
	}
}

func JsonRequest(r *http.Request, v interface{}) error {
	body, _ := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(body, v); err != nil {
		return err
	} else {
		return nil
	}
}

// 通过LDAP认证
func authLDAP(w http.ResponseWriter, r *http.Request) {

	// 改成支持json
	var Login struct {
		Username string
		Password string
	}
	code, msg, token := 0, "成功", ""
	if err := JsonRequest(r, &Login); err != nil {
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
	setTokenCookie(w, token)
	HTTPApiResult(w, TokenResult{Code: code, Message: msg, Token: token})
}

// setPrivateToken 设置私人token
func setPrivateToken(w http.ResponseWriter, r *http.Request) {
	code, msg := 0, "成功"

	var tenantPrivatetoken struct {
		Tenant string
		Token  string
	}

	JsonRequest(r, &tenantPrivatetoken)

	if tenantPrivatetoken.Tenant != "" && tenantPrivatetoken.Token != "" {
		if err := setAPPToken(r.Header.Get("Authorization-User"), tenantPrivatetoken.Tenant, tenantPrivatetoken.Token, 0); err != nil {
			code, msg = 10006, "写入redis失败"
		}
	} else {
		code, msg = 10005, "缺少必要的参数"
	}

	HTTPApiResult(w, JsonResult{Code: code, Message: msg})
}

// User 查询用户自身信息
func User(w http.ResponseWriter, r *http.Request) {
	result := UserResult{
		Code:    0,
		Message: "成功",
	}
	if sr, err := GS.LdapSeachUser(r.Header.Get("Authorization-User")); err == nil {
		result.Data = UserStruct{User: sr.GetAttributeValue("cn"), Mail: sr.GetAttributeValue("mail"),
			Name: sr.GetAttributeValue("name")}
	} else {
		result.Code, result.Message = 10009, "用户不存在"
	}
	HTTPApiResult(w, result)
}

// refreshToken token刷新
func refreshToken(w http.ResponseWriter, r *http.Request) {
	result := TokenResult{
		Code:    0,
		Message: "成功",
		Token:   "",
	}
	if claims, err := validateToken(r.Header.Get("Authorization")); err == nil {
		if signedToken, err := createToken(&Claims{Username: claims.Username, Password: claims.Password}); err != nil {
			result.Code, result.Message = 10003, err.Error()
		} else {
			result.Token = signedToken
		}
	}
	HTTPApiResult(w, result)
}

func init() {
	http.HandleFunc("/auth/refreshToken", refreshToken)
	http.HandleFunc("/auth/ldap", authLDAP)
	http.HandleFunc("/user/privateToken", GS.middleware(setPrivateToken))
	http.HandleFunc("/user", GS.middleware(User))
}
