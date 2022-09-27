package gateway

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

func MidBlack(writer http.ResponseWriter, request *http.Request) error {
	return nil
}

func (g *Gateway) middleware(handle Handle) Handle {
	return func(w http.ResponseWriter, r *http.Request) {
		// todo 黑名单
		if err := MidBlack(w, r); err != nil {
			return
		}
		// 监控指标变更
		httpRequestCounter.Inc()

		// 验证用户是否登录
		if token := r.Header.Get("Authorization"); token != "" {
			if claims, err := vaildateToken(token); err != nil {
				HTTPApiResult(w, JsonResult{Code: 10004, Message: "token验证失败"})
				UserRequestCounter.With(prometheus.Labels{"user": "guest"}).Inc()
				return
			} else {
				// 避免重复解码占用CPU
				r.Header.Set("Authorization-User", claims.Username)
				UserRequestCounter.With(prometheus.Labels{"user": claims.Username}).Inc()
			}
		} else {
			HTTPApiResult(w, JsonResult{Code: 10008, Message: "没有携带token信息"})
			UserRequestCounter.With(prometheus.Labels{"user": "guest"}).Inc()
			return
		}
		handle(w, r)
	}
}
