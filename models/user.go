package models

import (
	"goapi/utils"
	"database/sql"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/guregu/null"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

var (
	_ = time.Second
	_ = sql.LevelDefault
	_ = null.Bool{}
)

//User is a user !!!
type User struct {
	IDUser                uint            `gorm:"column:id_user;primary_key" json:"id_user"`
	IDLang                int             `gorm:"column:id_lang" json:"id_lang"`
	IDImage               int             `gorm:"column:id_image" json:"id_image"`
	Lastname              string          `gorm:"column:lastname" json:"lastname"`
	Firstname             string          `gorm:"column:firstname" json:"firstname"`
	Biography             null.String     `gorm:"column:biography" json:"biography"`
	Email                 string          `gorm:"column:email;unique_index" json:"email"`
	Passwd                string          `gorm:"column:passwd" json:"passwd"`
	LastPasswdGen         time.Time       `gorm:"column:last_passwd_gen" json:"last_passwd_gen"`
	StatsDateFrom         null.Time       `gorm:"column:stats_date_from" json:"stats_date_from"`
	StatsDateTo           null.Time       `gorm:"column:stats_date_to" json:"stats_date_to"`
	StatsCompareFrom      null.Time       `gorm:"column:stats_compare_from" json:"stats_compare_from"`
	StatsCompareTo        null.Time       `gorm:"column:stats_compare_to" json:"stats_compare_to"`
	StatsCompareOption    int             `gorm:"column:stats_compare_option" json:"stats_compare_option"`
	PreselectDateRange    null.String     `gorm:"column:preselect_date_range" json:"preselect_date_range"`
	BoColor               null.String     `gorm:"column:bo_color" json:"bo_color"`
	BoTheme               null.String     `gorm:"column:bo_theme" json:"bo_theme"`
	BoCSS                 null.String     `gorm:"column:bo_css" json:"bo_css"`
	DefaultTab            int             `gorm:"column:default_tab" json:"default_tab"`
	BoWidth               int             `gorm:"column:bo_width" json:"bo_width"`
	BoMenu                int             `gorm:"column:bo_menu" json:"bo_menu"`
	Active                int             `gorm:"column:active" json:"active"`
	Optin                 int             `gorm:"column:optin" json:"optin"`
	LastConnectionDate    null.Time       `gorm:"column:last_connection_date" json:"last_connection_date"`
	ResetPasswordToken    string          `gorm:"column:reset_password_token;type:varchar(64)" json:"reset_password_token"`
	ResetPasswordValidity null.Time       `gorm:"column:reset_password_validity" json:"reset_password_validity"`
	Accesses              utils.RightBits `gorm:"column:accesses" json:"accesses"`
	Token                 string          `gorm:"-" json:"token"`
	//Association
	Image *Object `gorm:"preload:true;save_associations:false;associations_autocreate:false;associations_autoupdate:false;foreignkey:id_image;AssociationForeignKey:id_object" json:"image,omitempty"`
}

// TableName sets the insert table name for this struct type
func (u *User) TableName() string {
	return "user"
}

//FilterColumns to return default columns to filter on
func (u *User) FilterColumns() map[string]string {
	return map[string]string{}
}

//OrderColumns return available order columns
func (u *User) OrderColumns() []string {
	return []string{"email", "lastname", "firstname"}
}

//Validate incoming user details...
func (u *User) Validate() (map[string]interface{}, bool) {

	if !strings.Contains(u.Email, "@") {
		return utils.Message(false, "Email address is required"), false
	}
	//Email must be unique
	temp := &User{}

	//check for errors and duplicate emails
	err := GetDB().Table(temp.TableName()).Where("email = ?", u.Email).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return utils.Message(false, "Connection error. Please retry"), false
	}
	if u.IDUser == 0 {
		//If new user then check passwd and email
		if len(u.Passwd) < 6 {
			return utils.Message(false, "Password is required"), false
		}
		if temp.Email != "" {
			return utils.Message(false, "Email address already in use by another user."), false
		}
		//Update new password with encryption !!
		u.NewPassword(u.Passwd)
	} else if u.IDUser > 0 {
		//Modify mail adress for another existing mail adress
		if temp.Email != "" && temp.IDUser != u.IDUser {
			return utils.Message(false, "Email address already in use by another user."), false
		}
		if len(u.Passwd) < 6 {
			return utils.Message(false, "Password is required"), false
		}
	}

	if u.Accesses == 0 {
		u.Accesses = utils.NoRight //default to no rights (being 0 :))
	}

	return utils.Message(false, "Requirement passed"), true
}

//NewPassword Change the password of the user
func (u *User) NewPassword(password string) {
	u.LastPasswdGen = time.Now() // set default time for lastgenpasswd
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u.Passwd = string(hashedPassword)
}

//GetUser return an user from its id
func GetUser(u uint) *User {
	user := &User{}
	GetDB().Set("gorm:auto_preload", true).Table("user").Where("id_user = ?", u).First(user)
	if user.Email == "" { //User not found!
		return nil
	}
	user.Passwd = ""
	return user
}

//LoginUser allows a user to login with email and password
func LoginUser(email, password string) map[string]interface{} {

	user := &User{}
	err := GetDB().Set("gorm:auto_preload", true).Table(user.TableName()).Where("email = ?", email).First(user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Message(false, "Invalid login credentials. Please try again")
		}
		return utils.Message(false, "Connection error. Please retry")
	}
	if user.Active == 0 {
		return utils.Message(false, "User hasn't been activated")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Passwd), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword { //Password does not match!
		return utils.Message(false, "Invalid login credentials. Please try again")
	}
	//Worked! Logged In
	user.Passwd = ""

	//Create JWT token
	//	tk := &utils.Token{
	//		UserId:       user.IDUser,
	//		UserRights:   user.Accesses,
	//		CompanyRight: user.GetCompanyRights(),
	//		TeamRight:    user.GetTeamRights(),
	//	}
	//	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	//	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	user.Token = user.GenToken() //Store the token in the response

	resp := utils.Message(true, "Logged In")
	resp["user"] = user
	return resp
}

//GenToken generate an auth token
func (u *User) GenToken() string {
	tk := &utils.Token{
		UserId:     u.IDUser,
		UserRights: u.Accesses,
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	return tokenString
}
