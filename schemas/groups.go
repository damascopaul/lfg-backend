package schemas

import (
	"errors"
	"fmt"
	"time"

	"github.com/damascopaul/lfg-backend/data"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type Group struct {
	ID          int64     `json:"id,omitempty" gorm:"primaryKey"`
	Title       string    `json:"title,omitempty" gorm:"not null"`
	Description string    `json:"description,omitempty"`
	Status      int16     `json:"status" gorm:"default:0"`
	Password    string    `json:"password,omitempty"`
	MaxSize     int16     `json:"max_size,omitempty" gorm:"default:5"`
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"autoCreateTime"`
	OwnerID     int64     `json:"owner_id" gorm:"not null"`
	Members     []User    `json:"members" gorm:"many2many:joined_groups"`

	DB *gorm.DB `json:"-" gorm:"-"`
}

func (g *Group) memberIndex(uid int64) int {
	return slices.IndexFunc(g.Members, func(m User) bool {
		return m.ID == uid
	})
}

// IsFull checks if the group is full.
func (g *Group) IsFull() bool {
	return g.MaxSize-1 == int16(len(g.Members))
}

// IsMember checks if the user is a member of the group.
func (g *Group) IsMember(uid int64) bool {
	// Return true if the user is in the member list of the group.
	i := g.memberIndex(uid)
	return i != -1
}

// IsOpen checks if the group is open.
func (g *Group) IsOpen() bool {
	return g.Status == 0
}

// IsOwner checks if the user is the owner of the group.
func (g *Group) IsOwner(uid int64) bool {
	return g.OwnerID == uid
}

// IsPrivate checks if the group is private.
func (g *Group) IsPrivate() bool {
	return g.Password != ""
}

// ValidatePassword validates the group password
func (g *Group) ValidatePassword(pw string) error {
	if g.Password != pw {
		// Return an error if the password does not match.
		log.Error("Password for group is invalid")
		return errors.New("incorrect group password")
	}
	return nil
}

// ValidateForCreate checks if the group is a valid new entry.
func (g *Group) ValidateForCreate() error {
	const FieldIsReqMsg string = "This field is required"
	var errors []FieldError

	const maxTitleLen int = 50
	if g.Title == "" {
		// Add a field error if the `title` field is empty
		errors = append(
			errors,
			FieldError{
				Name:  "title",
				Error: FieldIsReqMsg,
			})
	} else if len(g.Title) > maxTitleLen {
		// Add a field error if the `title` length is greater than 50
		errors = append(
			errors,
			FieldError{
				Name: "title",
				Error: fmt.Sprintf(
					"This field cannot be more than %v characters long", maxTitleLen),
			})
	}

	const maxDescLen int = 200
	if g.Description == "" {
		// Add a field error if the `description` field is empty
		errors = append(
			errors,
			FieldError{
				Name:  "description",
				Error: FieldIsReqMsg,
			})
	} else if len(g.Description) > maxDescLen {
		// Add a field error if the `description` length is greater than 200
		errors = append(
			errors,
			FieldError{
				Name: "description",
				Error: fmt.Sprintf(
					"This field cannot be more than %v characters long", maxDescLen),
			})
	}

	const (
		minSize int16 = 5
		maxSize int16 = 200
	)
	if g.MaxSize < minSize || g.MaxSize > maxSize {
		errors = append(
			errors,
			FieldError{
				Name: "max_size",
				Error: fmt.Sprintf(
					"The value should range from %v to %v", minSize, maxSize),
			})
	}

	log.Info("Validated new group request")
	if len(errors) > 0 {
		return &ValidationError{
			Message: "The new group is not valid",
			Errors:  errors,
		}
	}
	return nil
}

func preloadUser(db *gorm.DB) *gorm.DB {
	return db.Select("id", "username", "created_at")
}

func retrieveGroup(g *Group, fields []string) error {
	r := g.DB.Model(&g).Preload(
		"Members", preloadUser).Select(fields).First(&g, g.ID)
	if r.Error != nil {
		log.Fatalf("Could not retrieve group. Error: %v", r.Error.Error())
	} else {
		log.Info("Retrieved group successfully")
	}
	return r.Error
}

// InitDB initializes the database object
func (g *Group) InitDB() error {
	db, err := data.CreateConnection()
	if err != nil {
		return err
	}
	g.DB = db
	g.Migrate()
	log.WithFields(log.Fields{"model": "Group"}).Info("Initialized database")
	return nil
}

// Creates the group table based on the struct model
func (g *Group) Migrate() error {
	if err := g.DB.AutoMigrate(&g); err != nil {
		log.WithFields(
			log.Fields{"model": "Group"}).Fatal("Failed to auto migrate model")
		return err
	}
	log.WithFields(log.Fields{"model": "Group"}).Info("Auto migrated model")
	return nil
}

// Create adds a new group entry to the database.
func (g *Group) Create() error {
	r := g.DB.Create(&g)
	if r.Error != nil {
		log.Fatalf("Could not create group. Error: %v", r.Error.Error())
	} else {
		log.Info("Created group successfully")
	}
	return r.Error
}

// List gets all of the group entries from the database.
func (g *Group) List() ([]Group, error) {
	groups := []Group{}
	r := g.DB.Model(&g).Preload("Members", preloadUser).Select(
		"id", "title", "description", "status",
		"max_size", "created_at", "owner_id",
	).Find(&groups)
	if r.Error != nil {
		log.Fatalf("Could not list group. Error: %v", r.Error.Error())
	} else {
		log.Info("Listed groups successfully")
	}
	return groups, r.Error
}

// Retrieve retrieves the group details from the database given its database ID.
func (g *Group) Retrieve() error {
	fields := []string{
		"id", "title", "description",
		"status", "max_size", "created_at", "owner_id",
	}
	return retrieveGroup(g, fields)
}

// RetrieveWithPassword returns the group details from the database given its ID.
//
// The returned Group includes the password value.
func (g *Group) RetrieveWithPassword() error {
	fields := []string{
		"id", "title", "description", "password",
		"status", "max_size", "created_at", "owner_id",
	}
	return retrieveGroup(g, fields)
}

// Update updates a group entry.
func (g *Group) Update() error {
	r := g.DB.Save(&g)
	if r.Error != nil {
		log.Fatalf("Could not update group. Error: %v", r.Error.Error())
	} else {
		log.Info("Updated the group successfully")
	}
	return r.Error
}

// RemoveMember removes a user from the group.
func (g *Group) RemoveMember(u User) error {
	if err := g.DB.Model(&g).Association("Members").Delete(u); err != nil {
		log.Fatalf("Could not remove group member. Error: %v", err)
		return err
	}
	log.Info("Removed the member from the group successfully")
	return nil
}
