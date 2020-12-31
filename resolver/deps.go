package resolver

import (
	db "auth-grpc/db"
	pb "auth-grpc/proto"

	_ "github.com/go-sql-driver/mysql" // New import
)

//Server Inject dependencies here
type Server struct {
	pb.UnimplementedAuthServer
	//dependencies
	User *db.UserModel
}
