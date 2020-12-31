package db

import (
	pb "auth-grpc/proto"
	"database/sql"
	"errors"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

//UserModel access to dbs
type UserModel struct {
	DB *sql.DB
}

// func (m *UserModel) GetUser(id int64) (int64, error) {

// 	query := "SELECT uuid, email, username, active, created FROM user WHERE id = ?"

// 	row := m.DB.QueryRow(query, id)

// }

//CreateUser creates a new user given valid sign up details
func (m *UserModel) CreateUser(input *pb.SignUpData) (int64, error) {

	uuid, _ := uuid.NewRandom()

	//we assume that our data has already been validated inside our api handler call
	res, err := m.DB.Exec("INSERT INTO users (uuid, name, email, username, hashed_password, active, created) VALUES (?, ?, ?, ?, ?, ?, UTC_TIMESTAMP())", uuid.String(), input.Name, input.Email, input.Username, input.Password, 1)

	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			//check for duplicate field found

			if mySQLError.Number == 1062 {

				msg := mySQLError.Message

				if CheckError(msg, "username") {
					return 0, errors.New("Username '" + input.Username + "' is already taken")
				} else if CheckError(msg, "email") {
					return 0, errors.New("Email '" + input.Email + "' is already taken")
				}

			}
		}
		return 0, err
	}

	return res.LastInsertId()
}

//CheckError checks errors
func CheckError(msg, err string) bool {
	return strings.Contains(msg, err)

}

//EmailExists check if email exists
func (m *UserModel) EmailExists(email *pb.Email) bool {

	var uuid *string
	query := "SELECT uuid FROM users WHERE email = ?"
	row := m.DB.QueryRow(query, email)

	//check if error with query or no row returned
	err := row.Scan(&uuid)

	if err != nil {
		print("hello error")
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		return false

	}

	//if email exists, return true

	return true

}

//ResetPassword resets password for given user's email
func (m *UserModel) ResetPassword(newPass, newPassConfirm, email string) error {

	//check that passwords match
	if newPass != newPassConfirm {
		return errors.New("Passwords do not match")
	}

	//validate new password
	valid := m.ValidatePassword(newPass)

	if !valid {
		return errors.New("Password must be at least 6 characters long")
	}

	//hash new password for database update
	newPassHash, err := bcrypt.GenerateFromPassword([]byte(newPass), 12)

	if err != nil {
		return err
	}

	//update password hash in database
	query := "UPDATE USERS SET hashed_password = ? WHERE email = ?"

	_, err = m.DB.Exec(query, newPassHash, email)

	if err != nil {
		return err
	}

	return nil

}

//ValidatePassword validates password security
func (m *UserModel) ValidatePassword(password string) bool {

	// hash := make(map[int32]int)

	if len(password) < 6 {
		return false
	}

	return true
}

//Authenticate creates a new user given valid sign up details
func (m *UserModel) Authenticate(input *pb.LoginData) (int64, error) {

	// Retrieve the id and hashed password associated with the given email. If no
	// matching email exists, or the user is not active, we return the
	// ErrInvalidCredentials error.en
	var id int64
	var hashedPassword []byte
	stmt := "SELECT id, hashed_password FROM users WHERE username = ?"
	row := m.DB.QueryRow(stmt, input.Username)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("Invalid username or password")
		}
		return 0, err

	}
	// Check whether the hashed password and plain-text password provided match. // If they don't, we return the ErrInvalidCredentials error.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(input.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, errors.New("Invalid username or password")
		}
		return 0, err

	}
	// Otherwise, the password is correct. Return the user ID.
	return id, nil

}

//IsValidEmail checks for valid email format using web standard
func (m *UserModel) IsValidEmail(email string) bool {
	var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$]")
	err := m.matchesPattern(email, EmailRX)

	if err != nil {
		return false
	}
	return true
}

//MatchesPattern match pattern
func (m *UserModel) matchesPattern(field string, pattern *regexp.Regexp) error {
	value := field
	if value == "" {
		return nil
	}

	if !pattern.MatchString(value) {
		return nil
	}

	return nil
}
