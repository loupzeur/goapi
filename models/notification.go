package models

import (
	"bytes"
	"goapi/utils"
	"database/sql"
	"text/template"
	"time"

	"github.com/guregu/null"
)

var (
	_ = time.Second
	_ = sql.LevelDefault
	_ = null.Bool{}
)

const (
	NotificationNone = 0
)

const (
	NotificationMail = 1 << iota
	NotificationWeb
)

//Notification a notification
type Notification struct {
	IDNotification   uint      `gorm:"column:id_notification;primary_key" json:"id_notification"`
	IDUser           uint      `gorm:"column:id_user" json:"id_user"`
	Flag             uint      `gorm:"column:flag" json:"flag"` //Important, ...
	Title            string    `gorm:"column:title" json:"title"`
	Text             string    `gorm:"column:text" json:"text"`
	Type             uint      `gorm:"column:type" json:"type"` //Mail, Web, None
	Sent             bool      `gorm:"column:sent" json:"sent"`
	Seen             bool      `gorm:"column:seen" json:"seen"`
	DateNotification time.Time `gorm:"column:date_notification" json:"date_notification"`
	DateSent         null.Time `gorm:"column:date_sent" json:"date_sent"`
	DateSeen         null.Time `gorm:"column:date_seen" json:"date_seen"`
	//	To               string    `gorm:"column:to" json:"to"`
	Context utils.JSON `gorm:"column:context;type:json" json:"context,omitempty"`
	//Association
	User *User `gorm:"preload:false;save_associations:false;associations_autocreate:false;associations_autoupdate:false;foreignkey:id_user;AssociationForeignKey:id_user" json:"user,omitempty"`
}

// TableName sets the insert table name for this struct type
func (a *Notification) TableName() string {
	return "notification"
}

//Validate to validate a model
func (a *Notification) Validate() (map[string]interface{}, bool) {
	if a.IDNotification == 0 {
		a.DateNotification = time.Now()
	}
	return nil, true
}

//OrderColumns return available order columns
func (a *Notification) OrderColumns() []string {
	return []string{"date_notification", "date_sent"}
}

//FilterColumns to return default columns to filter on
func (a *Notification) FilterColumns() map[string]string {
	return map[string]string{"id_user": "int"}
}

//SentMail function to send mail
func (a *Notification) SentMail() {
	a.Sent = true
	a.DateSent = null.TimeFrom(time.Now())
	GetDB().Save(a)
}

//SendNotification insert a new notification in database
func SendNotification(title string, text string, to User, typeNotif uint, flag uint, variables utils.JSON) (*Notification, error) {
	t, _ := (template.New("notification")).Parse(text)
	var tpl bytes.Buffer
	t.Execute(&tpl, variables.ToInterface()) //Template at execution or sending ?
	notif := &Notification{
		Title:            title,
		Text:             tpl.String(), //text template + html template when sending by mail ?
		IDUser:           to.IDUser,
		Type:             typeNotif,
		Flag:             flag,
		DateNotification: time.Now(),
		Context:          variables,
	}
	if err := GetDB().Save(notif).Error; err != nil {
		return nil, err
	}
	return notif, nil
}
