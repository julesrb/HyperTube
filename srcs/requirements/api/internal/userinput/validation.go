package userinput

import (
	"regexp"
	"strings"
	"unicode"

	"hypertube/api/internal/i18n"
)

const (
	minPasswordBytes  = 8
	maxPasswordBytes  = 72
	minUsernameLength = 3
	maxUsernameLength = 32
	maxNameLength     = 100
)

var usernameCharsPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
var emailPrefixPattern = regexp.MustCompile(`^[A-Za-z0-9._+\-]+$`)
var emailDomainPattern = regexp.MustCompile(`^[A-Za-z0-9.\-]+$`)

var commonPasswords = map[string]struct{}{
	"00000000":      {},
	"11111111":      {},
	"12345678":      {},
	"123456789":     {},
	"abc123":        {},
	"admin":         {},
	"administrator": {},
	"baseball":      {},
	"batman":        {},
	"computer":      {},
	"default":       {},
	"dragon":        {},
	"football":      {},
	"hello":         {},
	"iloveyou":      {},
	"internet":      {},
	"letmein":       {},
	"login":         {},
	"master":        {},
	"monkey":        {},
	"password":      {},
	"password1":     {},
	"princess":      {},
	"qwerty":        {},
	"qwerty123":     {},
	"secret":        {},
	"sunshine":      {},
	"superman":      {},
	"trustno1":      {},
	"welcome":       {},
}

func ValidateEmail(raw string) (string, i18n.Message, bool) {
	email := strings.TrimSpace(raw)
	if email == "" {
		return "", i18n.MsgEmailRequired, false
	}

	normalizedEmail, ok := NormalizeEmail(email)
	if !ok {
		return "", i18n.MsgValidEmailRequired, false
	}
	return normalizedEmail, "", true
}

func NormalizeEmail(raw string) (string, bool) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", false
	}

	if !validEmail(email) {
		return "", false
	}

	return email, true
}

func ValidateUsername(raw string) (string, i18n.Message, bool) {
	username := strings.TrimSpace(raw)
	if username == "" {
		return "", i18n.MsgUsernameRequired, false
	}
	if len(username) < minUsernameLength {
		return "", i18n.MsgUsernameTooShort, false
	}
	if len(username) > maxUsernameLength {
		return "", i18n.MsgUsernameTooLong, false
	}
	if !usernameCharsPattern.MatchString(username) {
		return "", i18n.MsgUsernameInvalidChars, false
	}
	return username, "", true
}

func ValidateLoginIdentifier(raw string) (string, i18n.Message, bool) {
	if email, ok := NormalizeEmail(raw); ok {
		return email, "", true
	}

	username, validationMessage, ok := ValidateUsername(raw)
	if ok {
		return username, "", true
	}
	if strings.Contains(raw, "@") {
		return "", i18n.MsgValidEmailRequired, false
	}
	if validationMessage == i18n.MsgUsernameTooShort || validationMessage == i18n.MsgUsernameTooLong {
		return "", validationMessage, false
	}
	return "", i18n.MsgLoginInvalid, false
}

func ValidateRequiredPassword(password string) (i18n.Message, bool) {
	if password == "" {
		return i18n.MsgPasswordRequired, false
	}
	if len(password) < minPasswordBytes {
		return i18n.MsgPasswordTooShort, false
	}
	if len(password) > maxPasswordBytes {
		return i18n.MsgPasswordTooLong, false
	}
	if isCommonPassword(password) {
		return i18n.MsgPasswordTooCommon, false
	}
	return "", true
}

func ValidateLoginPassword(password string) (i18n.Message, bool) {
	if password == "" {
		return i18n.MsgPasswordRequired, false
	}
	if len(password) > maxPasswordBytes {
		return i18n.MsgPasswordTooLong, false
	}
	return "", true
}

func ValidateUpdatePassword(password string) (i18n.Message, bool) {
	if len(password) < minPasswordBytes {
		return i18n.MsgPasswordTooShort, false
	}
	if len(password) > maxPasswordBytes {
		return i18n.MsgPasswordTooLong, false
	}
	if isCommonPassword(password) {
		return i18n.MsgPasswordTooCommon, false
	}
	return "", true
}

func ValidateName(raw string, requiredMessage, tooLongMessage, invalidMessage i18n.Message) (string, i18n.Message, bool) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", requiredMessage, false
	}
	if len(name) > maxNameLength {
		return "", tooLongMessage, false
	}
	if !validPersonName(name) {
		return "", invalidMessage, false
	}
	return name, "", true
}

func validEmail(email string) bool {
	if strings.Count(email, "@") != 1 {
		return false
	}

	parts := strings.Split(email, "@")
	prefix, domain := parts[0], parts[1]
	if len(prefix) == 0 || len(prefix) > 64 {
		return false
	}
	if strings.HasPrefix(prefix, ".") || strings.HasSuffix(prefix, ".") || strings.Contains(prefix, "..") {
		return false
	}
	if !emailPrefixPattern.MatchString(prefix) {
		return false
	}

	if len(domain) == 0 || len(domain) > 253 {
		return false
	}
	if !emailDomainPattern.MatchString(domain) || !strings.Contains(domain, ".") {
		return false
	}

	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if label == "" || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
	}

	tld := labels[len(labels)-1]
	if len(tld) < 2 || len(tld) > 63 {
		return false
	}
	for _, r := range tld {
		if !unicode.IsLetter(r) || r > 127 {
			return false
		}
	}
	return true
}

func isCommonPassword(password string) bool {
	normalized := strings.ToLower(strings.TrimSpace(password))
	if _, ok := commonPasswords[normalized]; ok {
		return true
	}

	simplified := trimCommonPasswordSuffix(normalized)
	_, ok := commonPasswords[simplified]
	return ok
}

func trimCommonPasswordSuffix(password string) string {
	return strings.TrimRightFunc(password, func(r rune) bool {
		return unicode.IsDigit(r) || strings.ContainsRune("!?._-@#", r)
	})
}

func validPersonName(name string) bool {
	for _, r := range name {
		if unicode.IsLetter(r) || r == ' ' || r == '-' || r == '\'' {
			continue
		}
		return false
	}
	return true
}
