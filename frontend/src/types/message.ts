export const successMessages = {
    passwordChanged: "Votre mot de passe a bien été modifié.",
    emailChanged: "Votre adresse e-mail a été modifiée. Un e-mail de confirmation vient de vous être envoyé pour valider ce changement.",
    infoChanged: "Vos informations ont bien été mis à jour.",
};

export const errorMessages = {
    unknown: "Une erreur inattendue est survenue. Veuillez réessayer.",
    network: "Impossible de contacter le serveur. Vérifiez votre connexion internet.",
    unauthorized: "Vous devez être connecté pour effectuer cette action.",
    forbidden: "Vous n'avez pas les permissions nécessaires pour effectuer cette action.",

    passwordIncorrect: "Le mot de passe actuel est incorrect.",
    passwordTooShort: "Le mot de passe doit contenir au moins 8 caractères.",
    passwordTooWeak: "Le mot de passe est trop faible. Utilisez une combinaison de lettres, chiffres et caractères spéciaux.",
    passwordMismatch: "Les mots de passe ne correspondent pas.",
    passwordSameAsOld: "Le nouveau mot de passe doit être différent de l'ancien.",
    passwordChangeFailed: "Impossible de modifier le mot de passe. Veuillez réessayer.",

    emailInvalid: "L'adresse e-mail n'est pas valide.",
    emailAlreadyUsed: "Cette adresse e-mail est déjà utilisée.",
    emailChangeFailed: "Impossible de modifier l'adresse e-mail. Veuillez réessayer.",
    emailVerificationFailed: "Impossible d'envoyer l'e-mail de confirmation.",
    emailTokenInvalid: "Le lien de confirmation est invalide ou expiré.",

    usernameTooShort: "Le nom d'utilisateur est trop court.",
    usernameTooLong: "Le nom d'utilisateur est trop long.",
    usernameInvalid: "Le nom d'utilisateur contient des caractères invalides.",
    usernameAlreadyUsed: "Ce nom d'utilisateur est déjà utilisé.",
    usernameChangeFailed: "Impossible de modifier le nom d'utilisateur.",

    firstnameTooShort: "Le prénom est trop court.",
    firstnameTooLong: "Le prénom est trop long.",
    firstnameInvalid: "Le prénom contient des caractères invalides.",
    firstnameChangeFailed: "Impossible de modifier le prénom.",

    lastnameTooShort: "Le nom est trop court.",
    lastnameTooLong: "Le nom est trop long.",
    lastnameInvalid: "Le nom contient des caractères invalides.",
    lastnameChangeFailed: "Impossible de modifier le nom.",

    requiredField: "Ce champ est obligatoire.",
    requiredFields: "Veuillez remplir tous les champs obligatoires.",
    missingFields: "Certains champs obligatoires sont manquants. Veuillez vérifier le formulaire.",
};
