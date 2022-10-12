package schemas

import (
	"fmt"
	"time"

	"github.com/damascopaul/lfg-backend/data"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"unique"`
	Password     string    `json:"password,omitempty"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	MyGroups     []Group   `json:"-" gorm:"foreignKey:OwnerID"`
	JoinedGroups []Group   `json:"-" gorm:"many2many:joined_groups"`

	DB *gorm.DB `json:"-" gorm:"-"`
}

type TokenResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ValidateForSignUp checks if the user struct is valid for sign up.
func (u *User) ValidateForSignUp() error {
	const FieldIsReqMsg string = "This field is required"
	const (
		maxUsernameLen int = 50
		minPasswordLen int = 8
		maxPasswordLen int = 200
	)
	var errors []FieldError
	if u.Username == "" {
		// Add a field error if the `username` field is empty
		errors = append(
			errors,
			FieldError{
				Name:  "username",
				Error: FieldIsReqMsg,
			})
	} else if len(u.Username) > maxUsernameLen {
		// Add a field error if the `username` has more than 50 characters
		errors = append(
			errors,
			FieldError{
				Name: "username",
				Error: fmt.Sprintf(
					"This field cannot be more than %v characters long", maxUsernameLen),
			})
	}
	// TODO: Add more robust validation for username

	if u.Password == "" {
		// Add a field error if the `description` field is empty
		errors = append(
			errors,
			FieldError{
				Name:  "password",
				Error: FieldIsReqMsg,
			})
	} else if len(u.Password) < minPasswordLen ||
		len(u.Password) > maxPasswordLen {
		// Add a field error if the `password` has more than 8 characters
		errors = append(
			errors,
			FieldError{
				Name: "password",
				Error: fmt.Sprintf(
					"This field has to be %v to %v characters long",
					minPasswordLen, maxUsernameLen),
			})
	}
	// TODO: Add more robust validation for password

	if len(errors) > 0 {
		log.WithFields(log.Fields{"model": "User"}).Warn("Request body is invalid")
		return &ValidationError{
			Message: "The request body contains errors",
			Errors:  errors,
		}
	}
	log.WithFields(log.Fields{"model": "User"}).Info("Request body is valid")
	return nil
}

// InitDB initializes the database object
func (u *User) InitDB() error {
	db, err := data.CreateConnection()
	if err != nil {
		return err
	}
	u.DB = db
	u.Migrate()
	log.WithFields(log.Fields{"model": "User"}).Info("Initialized database")
	return nil
}

// Migrate creates the user table based on the struct model
func (u *User) Migrate() error {
	if err := u.DB.AutoMigrate(&u); err != nil {
		log.WithFields(
			log.Fields{"model": "User"}).Fatal("Failed to auto migrate model")
		return err
	}
	log.WithFields(log.Fields{"model": "User"}).Info("Auto migrated model")
	return nil
}

// BeforeCreate hashes the password of the user before adding it to the DB.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	hashedPw, err := bcrypt.GenerateFromPassword(
		[]byte(u.Password), bcrypt.MinCost)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Could not hash user password")
		return err
	}
	u.Password = string(hashedPw)
	log.Debug("Hashed user password")
	return nil
}

// Create adds a new user entry to the database.
func (u *User) Create() error {
	r := u.DB.Create(&u)
	if r.Error != nil {
		log.Fatalf("Could not create user. Error: %v", r.Error.Error())
	} else {
		log.Info("Created user successfully")
	}
	return r.Error
}

// Retrieve retrieves the user details given its database ID.
func (u *User) Retrieve() error {
	r := u.DB.Select("id", "username", "created_at").First(&u, u.ID)
	if r.Error != nil {
		log.Fatalf("Could not retrieve user. Error: %v", r.Error)
	} else {
		log.Info("Retrieved the user successfully")
	}
	return r.Error
}

// RetrieveUserByUsername retrieves a user details given its username.
func (u *User) RetrieveByUsername() error {
	r := u.DB.Where("username = ?", u.Username).First(&u)
	if r.Error != nil {
		log.Errorf("Could not retrieve user by username. Error: %v", r.Error)
	} else {
		log.Info("Retrieved the user successfully")
	}
	return r.Error
}
