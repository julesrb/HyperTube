package i18n

import (
	"fmt"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/language"
)

type Locale string

const (
	English Locale = "en"
	French  Locale = "fr"
	German  Locale = "de"
)

type Message string

const (
	MsgInvalidJSONBody                 Message = "invalid JSON body"
	MsgEmailRequired                   Message = "email is required"
	MsgValidEmailRequired              Message = "invalid email"
	MsgUsernameRequired                Message = "username is required"
	MsgUsernameTooShort                Message = "username is too short"
	MsgUsernameTooLong                 Message = "username is too long"
	MsgUsernameInvalidChars            Message = "username has invalid characters"
	MsgUsernameInvalid                 Message = "invalid username"
	MsgFirstNameRequired               Message = "first name is required"
	MsgFirstNameTooLong                Message = "first name is too long"
	MsgFirstNameInvalid                Message = "invalid first name"
	MsgLastNameRequired                Message = "last name is required"
	MsgLastNameTooLong                 Message = "last name is too long"
	MsgLastNameInvalid                 Message = "invalid last name"
	MsgPasswordTooShort                Message = "password is too short"
	MsgPasswordLength                  Message = "invalid password length"
	MsgLoginRequired                   Message = "email or username is required"
	MsgLoginInvalid                    Message = "invalid email or username"
	MsgPasswordRequired                Message = "password is required"
	MsgPasswordInvalid                 Message = "invalid password"
	MsgPasswordTooCommon               Message = "password is too common"
	MsgCurrentPasswordInvalid          Message = "current password is invalid"
	MsgNewPasswordSameAsCurrent        Message = "new password must differ from current password"
	MsgPasswordChangeSuccess           Message = "password has been changed"
	MsgFailedChangePassword            Message = "failed to change password"
	MsgPasswordConfirmationMismatch    Message = "password confirmation does not match"
	MsgEmailAlreadyInUse               Message = "email is already in use"
	MsgUsernameAlreadyInUse            Message = "username is already in use"
	MsgFailedCreateUser                Message = "failed to create user"
	MsgUserNotFound                    Message = "user not found"
	MsgUserUpdateForbidden             Message = "cannot update another user's profile"
	MsgOAuthEmailUpdateForbidden       Message = "OAuth users cannot change their email"
	MsgOAuthPasswordUpdateForbidden    Message = "OAuth users cannot change their password"
	MsgOAuthUsernameUpdateForbidden    Message = "OAuth users cannot change their username"
	MsgOAuthFirstNameUpdateForbidden   Message = "OAuth users cannot change their first name"
	MsgOAuthLastNameUpdateForbidden    Message = "OAuth users cannot change their last name"
	MsgProfilePictureUpdateForbidden   Message = "profile picture can only be removed"
	MsgFailedUpdateUser                Message = "failed to update user"
	MsgInvalidUserColor                Message = "invalid user color"
	MsgFailedUpdateUserColor           Message = "failed to update user color"
	MsgInvalidCredentials              Message = "invalid email, username, or password"
	MsgFailedLoadUser                  Message = "failed to load user"
	MsgFailedCreateToken               Message = "failed to create token"
	MsgAuthServiceUnavailable          Message = "authentication service is unavailable"
	MsgPasswordResetEmailNotConfigured Message = "password reset email is not configured"
	MsgFailedCreatePasswordResetToken  Message = "failed to create password reset token"
	MsgFailedStorePasswordResetToken   Message = "failed to store password reset token"
	MsgPasswordResetURLNotConfigured   Message = "password reset URL is not configured"
	MsgFailedSendPasswordResetEmail    Message = "failed to send password reset email"
	MsgInvalidPasswordResetToken       Message = "password reset link is invalid or expired"
	MsgFailedResetPassword             Message = "failed to reset password"
	MsgPasswordResetSuccess            Message = "password has been reset"
	MsgPasswordResetAccepted           Message = "if the email exists, a password reset link has been sent"
	MsgOAuthProviderNotConfigured      Message = "OAuth provider %s is not configured"
	MsgFailedCreateOAuthState          Message = "failed to create OAuth state"
	MsgFailedStartOAuth                Message = "failed to start %s OAuth"
	MsgInvalidOAuthState               Message = "invalid OAuth state"
	MsgOAuthDenied                     Message = "OAuth authorization was denied for %s"
	MsgFailedExchangeOAuthCode         Message = "failed to exchange %s authorization code"
	MsgFailedCreateOAuthUser           Message = "failed to create OAuth user"
	MsgOAuthApplicationNameRequired    Message = "OAuth application name is required"
	MsgOAuthApplicationNameTooLong     Message = "OAuth application name is too long"
	MsgOAuthApplicationNotFound        Message = "OAuth application not found"
	MsgFailedCreateOAuthApplication    Message = "failed to create OAuth application"
	MsgFailedLoadOAuthApplications     Message = "failed to load OAuth applications"
	MsgFailedUpdateOAuthApplication    Message = "failed to update OAuth application"
	MsgFailedDeleteOAuthApplication    Message = "failed to delete OAuth application"
	MsgFailedCreateOAuthCredentials    Message = "failed to create OAuth credentials"
	MsgInvalidFrontendAuthCallbackURL  Message = "invalid frontend auth callback URL"
	MsgFailedCreateAuthResponse        Message = "failed to create auth response"
	MsgGrantTypeRequired               Message = "grant type is required"
	MsgUnsupportedGrantType            Message = "only client_credentials grant type is supported"
	MsgUsernamePasswordRequired        Message = "username and password are required"
	MsgClientCredentialsRequired       Message = "client_id and client_secret are required"
	MsgInvalidClientCredentials        Message = "invalid client credentials"
	MsgPasswordTooLong                 Message = "password is too long"
	MsgInvalidUsernamePassword         Message = "invalid username or password"
	MsgInvalidContentType              Message = "invalid Content-Type"
	MsgInvalidFormBody                 Message = "invalid form body"
	MsgRequestBodyFormOrJSON           Message = "request body must be form encoded or JSON"
	MsgFailedCreateAccessToken         Message = "failed to create access token"
	MsgRefreshTokenRequired            Message = "refresh token is required"
	MsgInvalidRefreshToken             Message = "refresh token is invalid or expired"
	MsgMissingBearerToken              Message = "missing bearer token"
	MsgTokenExpired                    Message = "token expired"
	MsgInvalidBearerToken              Message = "invalid bearer token"
	MsgFailedLoadMovies                Message = "failed to load movies"
	MsgMissingUserContext              Message = "missing user context"
	MsgMovieNotFound                   Message = "movie not found"
	MsgFailedLoadMovie                 Message = "failed to load movie"
	MsgFailedFetchMovieDetails         Message = "failed to fetch movie details"
	MsgTitleQueryRequired              Message = "title query parameter is required"
	MsgFailedCheckSearchCache          Message = "failed to check search cache"
	MsgFailedLoadSearchResults         Message = "failed to load search results"
	MsgFailedSearchMovies              Message = "failed to search movies"
	MsgFailedStoreMovie                Message = "failed to store movie"
	MsgFailedStoreTorrent              Message = "failed to store torrent"
	MsgNoTrackerSource                 Message = "no tracker source found for this movie"
	MsgFailedLoadTrackerSource         Message = "failed to load tracker source"
	MsgNoComments                      Message = "no comments"
	MsgFailedAccessComments            Message = "failed to access comments"
	MsgInvalidRequestBody              Message = "invalid request body"
	MsgFailedCreateComment             Message = "failed to create comment"
	MsgCommentsNotFound                Message = "comments not found"
	MsgFailedLoadComments              Message = "failed to load comments"
	MsgCommentNotFound                 Message = "comment not found"
	MsgFailedLoadComment               Message = "failed to load comment"
	MsgFailedUpdateComment             Message = "failed to update comment"
	MsgFailedDeleteComment             Message = "failed to delete comment"
)

const (
	MsgOAuthApplicationRedirectURIRequired Message = "OAuth application redirect URI is required"
	MsgOAuthApplicationRedirectURITooLong  Message = "OAuth application redirect URI is too long"
	MsgOAuthApplicationRedirectURIInvalid  Message = "OAuth application redirect URI must be an absolute http or https URL without userinfo or fragment"
)

var matcher = language.NewMatcher([]language.Tag{
	language.English,
	language.French,
	language.German,
})

var translations = map[Locale]map[Message]string{
	French: {
		MsgInvalidJSONBody:                 "corps JSON invalide",
		MsgEmailRequired:                   "email requis",
		MsgValidEmailRequired:              "email invalide",
		MsgUsernameRequired:                "nom d'utilisateur requis",
		MsgUsernameTooShort:                "nom d'utilisateur trop court",
		MsgUsernameTooLong:                 "nom d'utilisateur trop long",
		MsgUsernameInvalidChars:            "nom d'utilisateur contient des caractères invalides",
		MsgUsernameInvalid:                 "nom d'utilisateur invalide",
		MsgFirstNameRequired:               "prénom requis",
		MsgFirstNameTooLong:                "prénom trop long",
		MsgFirstNameInvalid:                "prénom invalide",
		MsgLastNameRequired:                "nom requis",
		MsgLastNameTooLong:                 "nom trop long",
		MsgLastNameInvalid:                 "nom invalide",
		MsgPasswordTooShort:                "mot de passe trop court",
		MsgPasswordLength:                  "longueur du mot de passe invalide",
		MsgLoginRequired:                   "email ou nom d'utilisateur requis",
		MsgLoginInvalid:                    "email ou nom d'utilisateur invalide",
		MsgPasswordRequired:                "mot de passe requis",
		MsgPasswordInvalid:                 "mot de passe invalide",
		MsgPasswordTooCommon:               "mot de passe trop courant",
		MsgCurrentPasswordInvalid:          "mot de passe actuel invalide",
		MsgNewPasswordSameAsCurrent:        "le nouveau mot de passe doit être différent du mot de passe actuel",
		MsgPasswordChangeSuccess:           "mot de passe modifié",
		MsgFailedChangePassword:            "échec de la modification du mot de passe",
		MsgPasswordConfirmationMismatch:    "la confirmation du mot de passe ne correspond pas",
		MsgEmailAlreadyInUse:               "email déjà utilisé",
		MsgUsernameAlreadyInUse:            "nom d'utilisateur déjà utilisé",
		MsgFailedCreateUser:                "échec de la création de l'utilisateur",
		MsgUserNotFound:                    "utilisateur introuvable",
		MsgUserUpdateForbidden:             "impossible de modifier le profil d'un autre utilisateur",
		MsgOAuthEmailUpdateForbidden:       "les utilisateurs OAuth ne peuvent pas modifier leur email",
		MsgOAuthPasswordUpdateForbidden:    "les utilisateurs OAuth ne peuvent pas modifier leur mot de passe",
		MsgOAuthUsernameUpdateForbidden:    "les utilisateurs OAuth ne peuvent pas modifier leur nom d'utilisateur",
		MsgOAuthFirstNameUpdateForbidden:   "les utilisateurs OAuth ne peuvent pas modifier leur prénom",
		MsgOAuthLastNameUpdateForbidden:    "les utilisateurs OAuth ne peuvent pas modifier leur nom",
		MsgProfilePictureUpdateForbidden:   "la photo de profil peut seulement être supprimée",
		MsgFailedUpdateUser:                "échec de la mise à jour de l'utilisateur",
		MsgInvalidUserColor:                "couleur utilisateur invalide",
		MsgFailedUpdateUserColor:           "échec de la mise à jour de la couleur utilisateur",
		MsgInvalidCredentials:              "email, nom d'utilisateur ou mot de passe invalide",
		MsgFailedLoadUser:                  "échec du chargement de l'utilisateur",
		MsgFailedCreateToken:               "échec de la création du jeton",
		MsgAuthServiceUnavailable:          "service d'authentification indisponible",
		MsgPasswordResetEmailNotConfigured: "email de réinitialisation du mot de passe non configuré",
		MsgFailedCreatePasswordResetToken:  "échec de la création du jeton de réinitialisation du mot de passe",
		MsgFailedStorePasswordResetToken:   "échec de l'enregistrement du jeton de réinitialisation du mot de passe",
		MsgPasswordResetURLNotConfigured:   "URL de réinitialisation du mot de passe non configurée",
		MsgFailedSendPasswordResetEmail:    "échec de l'envoi de l'email de réinitialisation du mot de passe",
		MsgInvalidPasswordResetToken:       "le lien de réinitialisation du mot de passe est invalide ou expiré",
		MsgFailedResetPassword:             "échec de la réinitialisation du mot de passe",
		MsgPasswordResetSuccess:            "le mot de passe a été réinitialisé",
		MsgPasswordResetAccepted:           "si l'email existe, un lien de réinitialisation du mot de passe a été envoyé",
		MsgOAuthProviderNotConfigured:      "OAuth %s n'est pas configuré",
		MsgFailedCreateOAuthState:          "échec de la création de l'état OAuth",
		MsgFailedStartOAuth:                "échec du démarrage de OAuth %s",
		MsgInvalidOAuthState:               "état OAuth invalide",
		MsgOAuthDenied:                     "autorisation OAuth refusée pour %s",
		MsgFailedExchangeOAuthCode:         "échec de l'échange du code d'autorisation %s",
		MsgFailedCreateOAuthUser:           "échec de la création de l'utilisateur OAuth",
		MsgOAuthApplicationNameRequired:    "nom de l'application OAuth requis",
		MsgOAuthApplicationNameTooLong:     "nom de l'application OAuth trop long",
		MsgOAuthApplicationNotFound:        "application OAuth introuvable",
		MsgFailedCreateOAuthApplication:    "échec de la création de l'application OAuth",
		MsgFailedLoadOAuthApplications:     "échec du chargement des applications OAuth",
		MsgFailedUpdateOAuthApplication:    "échec de la mise à jour de l'application OAuth",
		MsgFailedDeleteOAuthApplication:    "échec de la suppression de l'application OAuth",
		MsgFailedCreateOAuthCredentials:    "échec de la création des identifiants OAuth",
		MsgInvalidFrontendAuthCallbackURL:  "URL de callback d'authentification frontend invalide",
		MsgFailedCreateAuthResponse:        "échec de la création de la réponse d'authentification",
		MsgGrantTypeRequired:               "type de grant requis",
		MsgUnsupportedGrantType:            "seul le grant type client_credentials est pris en charge",
		MsgUsernamePasswordRequired:        "nom d'utilisateur et mot de passe requis",
		MsgClientCredentialsRequired:       "client_id et client_secret sont requis",
		MsgInvalidClientCredentials:        "identifiants client invalides",
		MsgPasswordTooLong:                 "mot de passe trop long",
		MsgInvalidUsernamePassword:         "nom d'utilisateur ou mot de passe invalide",
		MsgInvalidContentType:              "Content-Type invalide",
		MsgInvalidFormBody:                 "corps de formulaire invalide",
		MsgRequestBodyFormOrJSON:           "le corps de la requête doit être encodé en formulaire ou en JSON",
		MsgFailedCreateAccessToken:         "échec de la création du jeton d'accès",
		MsgRefreshTokenRequired:            "jeton d'actualisation requis",
		MsgInvalidRefreshToken:             "jeton d'actualisation invalide ou expiré",
		MsgMissingBearerToken:              "jeton bearer manquant",
		MsgTokenExpired:                    "jeton expiré",
		MsgInvalidBearerToken:              "jeton bearer invalide",
		MsgFailedLoadMovies:                "échec du chargement des films",
		MsgMissingUserContext:              "contexte utilisateur manquant",
		MsgMovieNotFound:                   "film introuvable",
		MsgFailedLoadMovie:                 "échec du chargement du film",
		MsgFailedFetchMovieDetails:         "échec de la récupération des détails du film",
		MsgTitleQueryRequired:              "le paramètre de requête title est requis",
		MsgFailedCheckSearchCache:          "échec de la vérification du cache de recherche",
		MsgFailedLoadSearchResults:         "échec du chargement des résultats de recherche",
		MsgFailedSearchMovies:              "échec de la recherche de films",
		MsgFailedStoreMovie:                "échec de l'enregistrement du film",
		MsgFailedStoreTorrent:              "échec de l'enregistrement du torrent",
		MsgNoTrackerSource:                 "aucune source de tracker trouvée pour ce film",
		MsgFailedLoadTrackerSource:         "échec du chargement de la source tracker",
		MsgNoComments:                      "aucun commentaire",
		MsgFailedAccessComments:            "échec de l'accès aux commentaires",
		MsgInvalidRequestBody:              "corps de requête invalide",
		MsgFailedCreateComment:             "échec de la création du commentaire",
		MsgCommentsNotFound:                "commentaires introuvables",
		MsgFailedLoadComments:              "échec du chargement des commentaires",
		MsgCommentNotFound:                 "commentaire introuvable",
		MsgFailedLoadComment:               "échec du chargement du commentaire",
		MsgFailedUpdateComment:             "échec de la mise à jour du commentaire",
		MsgFailedDeleteComment:             "échec de la suppression du commentaire",

		MsgOAuthApplicationRedirectURIRequired: "URI de redirection de l'application OAuth requis",
		MsgOAuthApplicationRedirectURITooLong:  "URI de redirection de l'application OAuth trop long",
		MsgOAuthApplicationRedirectURIInvalid:  "l'URI de redirection de l'application OAuth doit être une URL absolue http ou https sans informations utilisateur ni fragment",
	},
	German: {
		MsgInvalidJSONBody:                 "ungültiger JSON-Body",
		MsgEmailRequired:                   "E-Mail ist erforderlich",
		MsgValidEmailRequired:              "E-Mail ist ungültig",
		MsgUsernameRequired:                "Benutzername ist erforderlich",
		MsgUsernameTooShort:                "Benutzername ist zu kurz",
		MsgUsernameTooLong:                 "Benutzername ist zu lang",
		MsgUsernameInvalidChars:            "Benutzername enthält ungültige Zeichen",
		MsgUsernameInvalid:                 "Benutzername ist ungültig",
		MsgFirstNameRequired:               "Vorname ist erforderlich",
		MsgFirstNameTooLong:                "Vorname ist zu lang",
		MsgFirstNameInvalid:                "Vorname ist ungültig",
		MsgLastNameRequired:                "Nachname ist erforderlich",
		MsgLastNameTooLong:                 "Nachname ist zu lang",
		MsgLastNameInvalid:                 "Nachname ist ungültig",
		MsgPasswordTooShort:                "Passwort ist zu kurz",
		MsgPasswordLength:                  "Passwortlänge ist ungültig",
		MsgLoginRequired:                   "E-Mail oder Benutzername ist erforderlich",
		MsgLoginInvalid:                    "E-Mail oder Benutzername ist ungültig",
		MsgPasswordRequired:                "Passwort ist erforderlich",
		MsgPasswordInvalid:                 "Passwort ist ungültig",
		MsgPasswordTooCommon:               "Passwort ist zu häufig",
		MsgCurrentPasswordInvalid:          "aktuelles Passwort ist ungültig",
		MsgNewPasswordSameAsCurrent:        "neues Passwort muss sich vom aktuellen Passwort unterscheiden",
		MsgPasswordChangeSuccess:           "Passwort wurde geändert",
		MsgFailedChangePassword:            "Passwort konnte nicht geändert werden",
		MsgPasswordConfirmationMismatch:    "Passwortbestätigung stimmt nicht überein",
		MsgEmailAlreadyInUse:               "E-Mail wird bereits verwendet",
		MsgUsernameAlreadyInUse:            "Benutzername wird bereits verwendet",
		MsgFailedCreateUser:                "Benutzer konnte nicht erstellt werden",
		MsgUserNotFound:                    "Benutzer nicht gefunden",
		MsgUserUpdateForbidden:             "das Profil eines anderen Benutzers kann nicht aktualisiert werden",
		MsgOAuthEmailUpdateForbidden:       "OAuth-Benutzer können ihre E-Mail nicht ändern",
		MsgOAuthPasswordUpdateForbidden:    "OAuth-Benutzer können ihr Passwort nicht ändern",
		MsgOAuthUsernameUpdateForbidden:    "OAuth-Benutzer können ihren Benutzernamen nicht ändern",
		MsgOAuthFirstNameUpdateForbidden:   "OAuth-Benutzer können ihren Vornamen nicht ändern",
		MsgOAuthLastNameUpdateForbidden:    "OAuth-Benutzer können ihren Nachnamen nicht ändern",
		MsgProfilePictureUpdateForbidden:   "Profilbild kann nur entfernt werden",
		MsgFailedUpdateUser:                "Benutzer konnte nicht aktualisiert werden",
		MsgInvalidUserColor:                "Benutzerfarbe ist ungültig",
		MsgFailedUpdateUserColor:           "Benutzerfarbe konnte nicht aktualisiert werden",
		MsgInvalidCredentials:              "E-Mail, Benutzername oder Passwort ist ungültig",
		MsgFailedLoadUser:                  "Benutzer konnte nicht geladen werden",
		MsgFailedCreateToken:               "Token konnte nicht erstellt werden",
		MsgAuthServiceUnavailable:          "Authentifizierungsdienst ist nicht verfügbar",
		MsgPasswordResetEmailNotConfigured: "E-Mail für Passwort-Zurücksetzung ist nicht konfiguriert",
		MsgFailedCreatePasswordResetToken:  "Token für Passwort-Zurücksetzung konnte nicht erstellt werden",
		MsgFailedStorePasswordResetToken:   "Token für Passwort-Zurücksetzung konnte nicht gespeichert werden",
		MsgPasswordResetURLNotConfigured:   "URL für Passwort-Zurücksetzung ist nicht konfiguriert",
		MsgFailedSendPasswordResetEmail:    "E-Mail zur Passwort-Zurücksetzung konnte nicht gesendet werden",
		MsgInvalidPasswordResetToken:       "Link zur Passwort-Zurücksetzung ist ungültig oder abgelaufen",
		MsgFailedResetPassword:             "Passwort konnte nicht zurückgesetzt werden",
		MsgPasswordResetSuccess:            "Passwort wurde zurückgesetzt",
		MsgPasswordResetAccepted:           "falls die E-Mail existiert, wurde ein Link zur Passwort-Zurücksetzung gesendet",
		MsgOAuthProviderNotConfigured:      "OAuth-Anbieter %s ist nicht konfiguriert",
		MsgFailedCreateOAuthState:          "OAuth-State konnte nicht erstellt werden",
		MsgFailedStartOAuth:                "%s OAuth konnte nicht gestartet werden",
		MsgInvalidOAuthState:               "ungültiger OAuth-State",
		MsgOAuthDenied:                     "OAuth-Autorisierung für %s wurde abgelehnt",
		MsgFailedExchangeOAuthCode:         "%s Autorisierungscode konnte nicht ausgetauscht werden",
		MsgFailedCreateOAuthUser:           "OAuth-Benutzer konnte nicht erstellt werden",
		MsgOAuthApplicationNameRequired:    "Name der OAuth-Anwendung ist erforderlich",
		MsgOAuthApplicationNameTooLong:     "Name der OAuth-Anwendung ist zu lang",
		MsgOAuthApplicationNotFound:        "OAuth-Anwendung nicht gefunden",
		MsgFailedCreateOAuthApplication:    "OAuth-Anwendung konnte nicht erstellt werden",
		MsgFailedLoadOAuthApplications:     "OAuth-Anwendungen konnten nicht geladen werden",
		MsgFailedUpdateOAuthApplication:    "OAuth-Anwendung konnte nicht aktualisiert werden",
		MsgFailedDeleteOAuthApplication:    "OAuth-Anwendung konnte nicht gelöscht werden",
		MsgFailedCreateOAuthCredentials:    "OAuth-Zugangsdaten konnten nicht erstellt werden",
		MsgInvalidFrontendAuthCallbackURL:  "Frontend-Auth-Callback-URL ist ungültig",
		MsgFailedCreateAuthResponse:        "Auth-Antwort konnte nicht erstellt werden",
		MsgGrantTypeRequired:               "Grant Type ist erforderlich",
		MsgUnsupportedGrantType:            "Nur der client_credentials Grant Type wird unterstützt",
		MsgUsernamePasswordRequired:        "Benutzername und Passwort sind erforderlich",
		MsgClientCredentialsRequired:       "client_id und client_secret sind erforderlich",
		MsgInvalidClientCredentials:        "Client-Zugangsdaten sind ungültig",
		MsgPasswordTooLong:                 "Passwort ist zu lang",
		MsgInvalidUsernamePassword:         "Benutzername oder Passwort ist ungültig",
		MsgInvalidContentType:              "Content-Type ist ungültig",
		MsgInvalidFormBody:                 "Formular-Body ist ungültig",
		MsgRequestBodyFormOrJSON:           "Request-Body muss formularcodiert oder JSON sein",
		MsgFailedCreateAccessToken:         "Access-Token konnte nicht erstellt werden",
		MsgRefreshTokenRequired:            "Refresh-Token ist erforderlich",
		MsgInvalidRefreshToken:             "Refresh-Token ist ungültig oder abgelaufen",
		MsgMissingBearerToken:              "Bearer-Token fehlt",
		MsgTokenExpired:                    "Token ist abgelaufen",
		MsgInvalidBearerToken:              "Bearer-Token ist ungültig",
		MsgFailedLoadMovies:                "Filme konnten nicht geladen werden",
		MsgMissingUserContext:              "Benutzerkontext fehlt",
		MsgMovieNotFound:                   "Film nicht gefunden",
		MsgFailedLoadMovie:                 "Film konnte nicht geladen werden",
		MsgFailedFetchMovieDetails:         "Filmdetails konnten nicht abgerufen werden",
		MsgTitleQueryRequired:              "der Query-Parameter title ist erforderlich",
		MsgFailedCheckSearchCache:          "Suchcache konnte nicht geprüft werden",
		MsgFailedLoadSearchResults:         "Suchergebnisse konnten nicht geladen werden",
		MsgFailedSearchMovies:              "Filmsuche fehlgeschlagen",
		MsgFailedStoreMovie:                "Film konnte nicht gespeichert werden",
		MsgFailedStoreTorrent:              "Torrent konnte nicht gespeichert werden",
		MsgNoTrackerSource:                 "keine Tracker-Quelle für diesen Film gefunden",
		MsgFailedLoadTrackerSource:         "Tracker-Quelle konnte nicht geladen werden",
		MsgNoComments:                      "keine Kommentare",
		MsgFailedAccessComments:            "Kommentare konnten nicht abgerufen werden",
		MsgInvalidRequestBody:              "ungültiger Request-Body",
		MsgFailedCreateComment:             "Kommentar konnte nicht erstellt werden",
		MsgCommentsNotFound:                "Kommentare nicht gefunden",
		MsgFailedLoadComments:              "Kommentare konnten nicht geladen werden",
		MsgCommentNotFound:                 "Kommentar nicht gefunden",
		MsgFailedLoadComment:               "Kommentar konnte nicht geladen werden",
		MsgFailedUpdateComment:             "Kommentar konnte nicht aktualisiert werden",
		MsgFailedDeleteComment:             "Kommentar konnte nicht gelöscht werden",

		MsgOAuthApplicationRedirectURIRequired: "Redirect-URI der OAuth-Anwendung ist erforderlich",
		MsgOAuthApplicationRedirectURITooLong:  "Redirect-URI der OAuth-Anwendung ist zu lang",
		MsgOAuthApplicationRedirectURIInvalid:  "Redirect-URI der OAuth-Anwendung muss eine absolute http- oder https-URL ohne Userinfo und Fragment sein",
	},
}

func FromRequest(r *http.Request) Locale {
	if r == nil {
		return English
	}
	return FromHeader(r.Header.Get("Accept-Language"))
}

func FromHeader(header string) Locale {
	header = strings.TrimSpace(header)
	if header == "" {
		return English
	}

	tags, _, err := language.ParseAcceptLanguage(header)
	if err == nil && len(tags) > 0 {
		tag, _, _ := matcher.Match(tags...)
		return fromTag(tag)
	}
	return FromValue(header)
}

func FromValue(value string) Locale {
	value = strings.TrimSpace(value)
	if value == "" {
		return English
	}

	tag, err := language.Parse(value)
	if err == nil {
		return fromTag(tag)
	}

	lower := strings.ToLower(value)
	if len(lower) >= 2 {
		switch lower[:2] {
		case "fr":
			return French
		case "de":
			return German
		case "en":
			return English
		}
	}
	return English
}

func (l Locale) String() string {
	switch l {
	case French, German:
		return string(l)
	default:
		return string(English)
	}
}

func T(locale Locale, message Message, args ...any) string {
	template := string(message)
	if localeTranslations, ok := translations[locale]; ok {
		if translated, ok := localeTranslations[message]; ok {
			template = translated
		}
	}
	if len(args) == 0 {
		return capitalizeFirst(template)
	}
	return capitalizeFirst(fmt.Sprintf(template, args...))
}

func capitalizeFirst(value string) string {
	first, size := utf8.DecodeRuneInString(value)
	if first == utf8.RuneError && size == 0 {
		return value
	}

	upper := unicode.ToUpper(first)
	if upper == first {
		return value
	}
	return string(upper) + value[size:]
}

func fromTag(tag language.Tag) Locale {
	base, _ := tag.Base()
	switch base.String() {
	case "fr":
		return French
	case "de":
		return German
	default:
		return English
	}
}
