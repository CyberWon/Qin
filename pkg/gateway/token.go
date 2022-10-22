package gateway

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gomodule/redigo/redis"
	"net/http"
	"time"
)

type Claims struct {
	Username string `json:"username"`
	Password string `json:"password"`
	jwt.StandardClaims
}

func createToken(claims *Claims) (signedToken string, err error) {
	claims.ExpiresAt = time.Now().Add(time.Duration(GS.Config.Auth.Expire) * time.Second).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if signedToken, err := token.SignedString([]byte(GS.Config.Auth.Secret)); err != nil {
		log.Panicln("生成token失败", err.Error())
		return "", err
	} else {
		return signedToken, nil
	}
}

func validateToken(signedToken string) (c *Claims, err error) {

	if token, err := jwt.ParseWithClaims(signedToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("unexpected login method %v", token.Header["alg"])
			return nil, err
		}
		return []byte(GS.Config.Auth.Secret), nil
	}); err != nil {
		return nil, err
	} else {
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			return claims, nil
		} else {
			return nil, jwt.ErrInvalidKey
		}

	}
}
func setAPPToken(user, app, token string, ttl int) error {
	if _, err := GS.Cache.Do("SET",
		fmt.Sprintf("gs:%s:%s", user, app),
		token,
	); err != nil {
		return err
	} else {
		if ttl != 0 {
			if _, err := GS.Cache.Do("EXPIRE", fmt.Sprintf("gs:%s:%s", user, app), ttl); err != nil {
				log.Println("设置ttl失败", err)
			}
		}
	}
	return nil
}

func getAPPToken(user, app string) (token string, err error) {
	if replay, err := GS.Cache.Do("GET", fmt.Sprintf("gs:%s:%s", user, app)); err != nil {
		log.Error(err)
		return "", err
	} else {
		token, err = redis.String(replay, err)
		return token, err
	}
}

func delAPPToken(user, app string) error {
	if _, err := GS.Cache.Do("DEL",
		fmt.Sprintf("gs:%s:%s", user, app),
	); err != nil {
		return err
	}
	return nil
}

func setTokenCookie(w http.ResponseWriter, token string) {
	cookie := http.Cookie{
		Name:     "qin-token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   GS.Config.Auth.Expire,
	}
	http.SetCookie(w, &cookie)
	log.Println(w.Header())

}
