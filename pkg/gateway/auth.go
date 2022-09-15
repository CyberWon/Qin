package gateway

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"log"
)

func (g *Gateway) ldapConn() {

	addr := fmt.Sprintf("%s:%d", g.Config.Ldap.Host, g.Config.Ldap.Port)

	if conn, err := ldap.Dial("tcp", addr); err != nil {
		log.Println("ldap链接创建失败", err)
	} else {
		g.LDAP.Conn = conn
	}

	if err := g.LDAP.Conn.Bind(g.Config.Ldap.BindDn, g.Config.Ldap.Password); err != nil {
		log.Println("ldap认证错误，用户名或者密码不对", err)
	}

}

func (g *Gateway) LdapAuth(user, password string) error {
	filter := fmt.Sprintf("(&(objectClass=inetOrgPerson)(uid=%s))", user)
	searchRequest := ldap.NewSearchRequest(
		g.Config.Ldap.UserBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		//fmt.Sprintf("(&(objectClass=organizationalUnit))"),
		filter,
		[]string{"cn"},
		nil,
	)
	// 避免因为长时间没有与LDAP Server通信，导致连接断开。
	if g.LDAP.Conn.IsClosing() {
		g.ldapConn()
	}
	sr, err := g.LDAP.Conn.Search(searchRequest)
	if err != nil {
		log.Println("查询LDAP用户失败，")
		return err
	}

	// 2. 进行二次bind，验证用户pass是否正确
	userDN := sr.Entries[0].DN
	err = g.LDAP.Conn.Bind(userDN, password)
	if err != nil {
		log.Println("LDAP用户登录失败")
		return err
	}
	return nil
}
