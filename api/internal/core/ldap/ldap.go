package ldap

import (
	"fmt"
	"time"

	"github.com/apisix/manager-api/internal/conf"
	"github.com/apisix/manager-api/internal/log"
	ldap_v3 "github.com/go-ldap/ldap/v3"
)

var (
	l *ldap_v3.Conn
)

func Init() {
	// TODO implement ldap connection with TLS
	var err error
	l, err = ldap_v3.Dial("tcp", conf.LdapConfig.Host)
	if err != nil {
		log.Errorf("ldap connect error: %s", err)
	}
	err = l.Bind(conf.LdapConfig.BindDN, conf.LdapConfig.BindPassword)
	if err != nil {
		log.Error("ldap bind failed, user or password is wrong")
	}
}

func UserAuthentication(username, password string) bool {
	searchRequest := ldap_v3.NewSearchRequest(
		conf.LdapConfig.BaseDN,
		ldap_v3.ScopeWholeSubtree,
		ldap_v3.NeverDerefAliases,
		0,
		int(30*time.Second),
		false,
		fmt.Sprintf(conf.LdapFilter, username),
		[]string{"cn"},
		nil)
	searchResult, err := l.Search(searchRequest)
	if err != nil {
		log.Error(err.Error())
	}
	if len(searchResult.Entries) > 0 {
		userDN := searchResult.Entries[0].DN
		err := l.Bind(userDN, password)
		if err != nil {
			log.Error(err.Error())
			return false
		}
		return true
	}
	return false
}
