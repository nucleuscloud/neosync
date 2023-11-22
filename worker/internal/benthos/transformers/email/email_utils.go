package transformers_email

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
)

var tld = []string{"com", "org", "net", "edu", "gov", "app", "dev"}

func parseEmail(email string) ([]string, error) {

	inputEmail, err := mail.ParseAddress(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email 5322format: %s", email)
	}

	parsedEmail := strings.Split(inputEmail.Address, "@")

	return parsedEmail, nil
}

func IsValidEmail(email string) bool {
	// Regular expression pattern for a simple email validation
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(emailPattern)
	return regex.MatchString(email)
}

func IsValidDomain(domain string) bool {
	pattern := `^@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	rfcRegex, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	return rfcRegex.MatchString(domain)
}

func IsValidUsername(username string) bool {

	// Regex to match RFC 5322 username
	// Chars allowed: a-z A-Z 0-9 . - _
	// First char must be alphanumeric
	// Last char must be alphanumeric or numeric
	// 63 max chars
	rfcRegex := `^[A-Za-z0-9](?:[A-Za-z0-9!#$%&'*+-/=?^_` +
		`{|}~.]{0,62}[A-Za-z0-9])?$`

	matched, _ := regexp.MatchString(rfcRegex, username)

	return matched
}
