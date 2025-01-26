package kerror

import (
	"github.com/dlclark/regexp2"
)

const (
	EmailRegexPattern    = `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
	PhoneRegexPattern    = `^\+?[1-9]\d{10,15}$`
	PinRegexPattern      = `^\d{4,12}$`
	PasswordRegexPattern = `^(?=.*[A-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[*@$!%*#?&])\S{8,}$`
)

var (
	RegexpEmail   = regexp2.MustCompile(EmailRegexPattern, 0)
	RegexpPhone   = regexp2.MustCompile(PhoneRegexPattern, 0)
	RegexPassword = regexp2.MustCompile(PasswordRegexPattern, 0)
	RegexPin      = regexp2.MustCompile(PinRegexPattern, 0)
)
