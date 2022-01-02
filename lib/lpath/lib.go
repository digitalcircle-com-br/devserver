package lpath

import (
	gopath "path"
	"strings"

	"github.com/digitalcircle-com-br/devserver/lib/config"
)

func Resolve(p ...string) string {

	ret := ""
	for _, v := range p {
		v = strings.Replace(v, "~DS", config.Wd, 1)
		v = strings.Replace(v, "~", config.UserHome, 1)
		ret = gopath.Join(ret, v)
	}
	return ret
}
