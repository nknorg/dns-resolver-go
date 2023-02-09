package dnsresolver

import "regexp"

const pattern = `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`

var reg *regexp.Regexp

func init() {
	reg = regexp.MustCompile(pattern)
}

func IsEmail(email string) bool {
	return reg.MatchString(email)
}
