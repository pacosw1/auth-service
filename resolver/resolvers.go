package resolver

import (
	pb "auth-grpc/proto"
	"context"
	"errors"

	_ "github.com/go-sql-driver/mysql" // New import
	"golang.org/x/crypto/bcrypt"
)

//CreateUser creates a user

//ResetPassword resets password
func (s *Server) ResetPassword(ctx context.Context, input *pb.PasswordReset) (*pb.Response, error) {

	//prepare protobuf response
	resp := &pb.Response{
		Message: "Password reset successfully",
	}

	err := s.User.ResetPassword(input.NewPassword, input.Confirm, input.Email)

	if err != nil {
		return nil, err
	}

	return resp, nil

}

//CreateUser creates a user in database
func (s *Server) CreateUser(ctx context.Context, input *pb.SignUpData) (*pb.NewUser, error) {

	//validate email
	validEmail := s.User.IsValidEmail(input.Email)

	if !validEmail {
		return nil, errors.New("Invalid Email")
	}

	//validate password length
	valid := s.User.ValidatePassword(input.Password)

	if !valid {
		return nil, errors.New("Invalid Password, Must be at least etc")
	}

	//hash password if valid
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)

	if err != nil {
		return nil, err
	}

	//update password input
	input.Password = string(hash)

	userID, err := s.User.CreateUser(input)

	if err != nil {
		return nil, err
	}

	resp := &pb.NewUser{
		Id: userID,
	}

	return resp, nil

}

//VerifyEmail verifies email format
func (s *Server) VerifyEmail(ctx context.Context, input *pb.Email) (*pb.Response, error) {

	valid := s.User.IsValidEmail(input.Email)

	if !valid {
		return nil, errors.New("Invalid Email")
	}

	pbResp := &pb.Response{
		Message: "Valid Email",
	}

	return pbResp, nil
}

//Authenticate authenticates user credentials
func (s *Server) Authenticate(ctx context.Context, input *pb.LoginData) (*pb.AuthResponse, error) {

	userID, err := s.User.Authenticate(input)

	if err != nil {
		return nil, err
	}

	//prepare pb response
	pbResonse := &pb.AuthResponse{
		Id: userID,
	}

	return pbResonse, nil
}

// func (s *Server) LoginUser(input *pb.LoginData)

// rpc LogoutUser(Email) returns (Response) {};
//     //create user will authenticate automatically
//     rpc CreateUser(SignUpData) returns (LoginResponse) {};
//     rpc LoginUser(LoginData) returns (LoginResponse) {};
//     rpc ResetPassword(Email) returns (Response) {};
