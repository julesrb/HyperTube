package i18n

import "testing"

func TestFromHeaderMatchesSupportedLanguages(t *testing.T) {
	tests := []struct {
		header string
		want   Locale
	}{
		{header: "en", want: English},
		{header: "en-US,en;q=0.9", want: English},
		{header: "fr", want: French},
		{header: "de-DE,de;q=0.9,en;q=0.8", want: German},
		{header: "fr-CA,fr;q=0.9,en;q=0.8", want: French},
		{header: "es", want: English},
		{header: "not a language", want: English},
		{header: "", want: English},
	}

	for _, tt := range tests {
		if got := FromHeader(tt.header); got != tt.want {
			t.Fatalf("FromHeader(%q) = %q, want %q", tt.header, got, tt.want)
		}
	}
}

func TestTranslateMessages(t *testing.T) {
	if got := T(French, MsgValidEmailRequired); got != "Email invalide" {
		t.Fatalf("unexpected French message: %q", got)
	}
	if got := T(German, MsgInvalidCredentials); got != "E-Mail, Benutzername oder Passwort ist ungültig" {
		t.Fatalf("unexpected German message: %q", got)
	}
	if got := T(English, MsgOAuthProviderNotConfigured, "GitHub"); got != "OAuth provider GitHub is not configured" {
		t.Fatalf("unexpected English message: %q", got)
	}
	if got := T(French, MsgFailedCreateUser); got != "Échec de la création de l'utilisateur" {
		t.Fatalf("unexpected capitalized French message: %q", got)
	}
	if got := T(German, MsgProfilePictureUpdateForbidden); got != "Profilbild kann nur entfernt werden" {
		t.Fatalf("unexpected German profile picture message: %q", got)
	}
	if got := T(English, MsgRefreshTokenRequired); got != "Refresh token is required" {
		t.Fatalf("unexpected English refresh-token message: %q", got)
	}
	if got := T(French, MsgInvalidRefreshToken); got != "Jeton d'actualisation invalide ou expiré" {
		t.Fatalf("unexpected French refresh-token message: %q", got)
	}
	if got := T(German, MsgInvalidRefreshToken); got != "Refresh-Token ist ungültig oder abgelaufen" {
		t.Fatalf("unexpected German refresh-token message: %q", got)
	}

	tests := []struct {
		locale  Locale
		message Message
		want    string
	}{
		{English, MsgCurrentPasswordInvalid, "Current password is invalid"},
		{French, MsgCurrentPasswordInvalid, "Mot de passe actuel invalide"},
		{German, MsgCurrentPasswordInvalid, "Aktuelles Passwort ist ungültig"},
		{English, MsgNewPasswordSameAsCurrent, "New password must differ from current password"},
		{French, MsgNewPasswordSameAsCurrent, "Le nouveau mot de passe doit être différent du mot de passe actuel"},
		{German, MsgNewPasswordSameAsCurrent, "Neues Passwort muss sich vom aktuellen Passwort unterscheiden"},
		{English, MsgPasswordChangeSuccess, "Password has been changed"},
		{French, MsgPasswordChangeSuccess, "Mot de passe modifié"},
		{German, MsgPasswordChangeSuccess, "Passwort wurde geändert"},
		{English, MsgFailedChangePassword, "Failed to change password"},
		{French, MsgFailedChangePassword, "Échec de la modification du mot de passe"},
		{German, MsgFailedChangePassword, "Passwort konnte nicht geändert werden"},
		{English, MsgPasswordConfirmationMismatch, "Password confirmation does not match"},
		{French, MsgPasswordConfirmationMismatch, "La confirmation du mot de passe ne correspond pas"},
		{German, MsgPasswordConfirmationMismatch, "Passwortbestätigung stimmt nicht überein"},
	}

	for _, tt := range tests {
		if got := T(tt.locale, tt.message); got != tt.want {
			t.Errorf("T(%q, %q) = %q, want %q", tt.locale, tt.message, got, tt.want)
		}
	}
}
