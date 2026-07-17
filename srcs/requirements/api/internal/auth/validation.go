package auth

import (
	"strings"

	"hypertube/api/internal/i18n"
	"hypertube/api/internal/userinput"
)

type validationErrors map[string]i18n.Message

func validateRegisterRequest(req registerRequest) (CreateUserParams, validationErrors, bool) {
	fields := validationErrors{}

	email, validationMessage, ok := userinput.ValidateEmail(req.Email)
	if !ok {
		fields["email"] = validationMessage
	}

	username, validationMessage, ok := userinput.ValidateUsername(req.Username)
	if !ok {
		fields["username"] = validationMessage
	}

	firstName := strings.TrimSpace(req.FirstName)
	if firstName == "" {
		firstName = strings.TrimSpace(req.FrontendFirstName)
	}
	if name, validationMessage, ok := userinput.ValidateName(firstName, i18n.MsgFirstNameRequired, i18n.MsgFirstNameTooLong, i18n.MsgFirstNameInvalid); ok {
		firstName = name
	} else {
		fields["first_name"] = validationMessage
	}

	lastName := strings.TrimSpace(req.LastName)
	if lastName == "" {
		lastName = strings.TrimSpace(req.FrontendLastName)
	}
	if name, validationMessage, ok := userinput.ValidateName(lastName, i18n.MsgLastNameRequired, i18n.MsgLastNameTooLong, i18n.MsgLastNameInvalid); ok {
		lastName = name
	} else {
		fields["last_name"] = validationMessage
	}

	if validationMessage, ok := userinput.ValidateRequiredPassword(req.Password); !ok {
		fields["password"] = validationMessage
	}

	if len(fields) > 0 {
		return CreateUserParams{}, fields, false
	}

	return CreateUserParams{
		Email:     email,
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
	}, nil, true
}

func validateLoginRequest(req loginRequest) (string, validationErrors, bool) {
	fields := validationErrors{}

	login := strings.TrimSpace(req.Login)
	if login == "" {
		fields["login"] = i18n.MsgLoginRequired
	} else {
		identifier, validationMessage, ok := userinput.ValidateLoginIdentifier(login)
		if ok {
			login = identifier
		} else {
			fields["login"] = validationMessage
		}
	}
	if validationMessage, ok := userinput.ValidateLoginPassword(req.Password); !ok {
		fields["password"] = validationMessage
	}

	if len(fields) > 0 {
		return "", fields, false
	}
	return login, nil, true
}
