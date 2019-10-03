package utils

//Notification texts and title for notifications
type Notification struct {
	Title string
	Text  string
}

const (
	NotificationUserCreation      int = 0
	NotificationProjectCreation       = 1
	NotificationUserPasswordReset     = 2
	NotificationTeamInvitation        = 3
	NotificationTeamProjet            = 4
	NotificationSupportCreation       = 5
)

//Notifications Array of existing notifications
var Notifications = map[int]Notification{
	NotificationUserCreation: Notification{
		Title: "Création de votre compte",
		Text:  "Bonjour, \nvotre compte a été crée, merci de le valider en visitant sur le lien suivant : {{.Link}}",
	},
	NotificationUserPasswordReset: Notification{
		Title: "Reset de votre mot de passe",
		Text:  "Bonjour, \nUne demande de mise à zéro de votre mot de passe a été demandé, pour le changer vous pouvez visiter le lien suivant : {{.Link}}",
	},
}
