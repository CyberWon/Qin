package gateway

import (
	"encoding/json"
	"net/http"
)

func (g *Gateway) middleware(handle Handle) Handle {
	return func(w http.ResponseWriter, r *http.Request) {
		// todo: 黑名单

		// 验证用户是否登录
		if token := r.Header.Get("Authorization"); token != "" {
			if claims, err := vaildateToken(token); err != nil {
				data, _ := json.Marshal(JsonResult{Code: 10004, Message: "token验证失败"})
				w.Write(data)
				return
			} else {
				// 避免重复解码占用CPU
				r.Header.Set("Authorization-User", claims.Username)
			}
		} else {
			data, _ := json.Marshal(JsonResult{Code: 10008, Message: "没有携带token信息"})
			w.Write(data)
			return
		}
		handle(w, r)
	}
}
