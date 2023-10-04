package main

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	_ "github.com/benthosdev/benthos/v4/public/components/io"
	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/bxcodec/faker/v4"
)

type EmailTransformerOptions struct {
	PreserveLength bool
	PreserveDomain bool
}

// main plugin logic goes here
func (e *EmailTransformerOptions) Process(ctx context.Context, m *service.Message) (service.MessageBatch, error) {

	b, err := m.AsBytes()
	if err != nil {
		return nil, err
	}

	parsedEmail, err := parseEmail(b)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %s", string(b))
	}

	var email string

	if e.PreserveDomain && !e.PreserveLength {

		email = faker.Username() + "@" + parsedEmail[1]

	} else if e.PreserveLength && !e.PreserveDomain {

		//preserve length of email but not the domain

		splitDomain := strings.Split(parsedEmail[1], ".") //split the domain to account for different domain name lengths

		domain := sliceString(faker.DomainName(), len(splitDomain[0]))

		tld := sliceString(faker.DomainName(), len(splitDomain[1]))

		email = sliceString(faker.Username(), len(parsedEmail[0])) + "@" + domain + "." + tld

	} else if e.PreserveDomain && e.PreserveLength {

		//preserve the domain and the length of the email -> keep the domain the same but slice the username to be the same length as the input username
		unLength := len(parsedEmail[0])

		un := faker.Username()

		email = sliceString(un, unLength) + "@" + parsedEmail[1]

	} else {
		//generate random email

		email = faker.Email()
	}

	m.SetBytes([]byte(email))

	return service.MessageBatch{m}, nil
}

// gets called when benthos is done with the component
func (e *EmailTransformerOptions) Close(ctx context.Context) error {
	return nil
}

func main() {

	spec := service.NewConfigSpec().
		Field(service.NewBoolField("preserve_domain")).
		Field(service.NewBoolField("preserve_length"))

	constructor := func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {

		//extract the params
		pd, err := conf.FieldBool("preserve_domain")
		if err != nil {
			return nil, err
		}

		pl, err := conf.FieldBool("preserve_length")
		if err != nil {
			return nil, err
		}
		return &EmailTransformerOptions{PreserveLength: pl, PreserveDomain: pd}, nil

	}

	//register the plugin
	service.RegisterProcessor("emailtransformer", spec, constructor)

	//executes the plugin
	service.RunCLI(context.Background())

}

func parseEmail(email []byte) ([]string, error) {

	inputEmail, err := mail.ParseAddress(string(email))
	if err != nil {

		return nil, fmt.Errorf("invalid email format: %s", email)
	}

	parsedEmail := strings.Split(inputEmail.Address, "@")

	return parsedEmail, nil
}

func sliceString(s string, l int) string {

	runes := []rune(s) //use runes instead of strings in order to avoid slicing a multi-byte character and returning invalid UTF-8

	if l > len(runes) {
		l = len(runes)
	}

	return string(runes[:l])
}
