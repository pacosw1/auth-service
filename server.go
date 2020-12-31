package main

import (
	db "auth-grpc/db"
	pb "auth-grpc/proto"
	"auth-grpc/resolver"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	_ "github.com/go-sql-driver/mysql" // New import
	"github.com/gomodule/redigo/redis"

	"google.golang.org/grpc"
)

const (
	port = ":9000"
)

//MySQLServer server
type MySQLServer struct {
	connected bool
	dsn       string
	// dialTimeout time.Duration
	maxRetries int
	conn       *sql.DB
}

func main() {

	//wait for db to start
	// time.Sleep(6 * time.Second)

	//init sql server
	sqlServ := &MySQLServer{
		dsn:        "root:root@tcp(localhost:4100)/monolith_db?timeout=3s&parseTime=true",
		connected:  false,
		maxRetries: 10,
	}

	//init redis server
	redisServ := &RedisServer{
		host:        "localhost",
		port:        "4200",
		maxRetries:  10,
		connected:   false,
		dialTimeout: 3 * time.Second,
	}

	//init connection manager
	connManager := &ConnectionManager{
		Servers:   []ExternalServer{sqlServ, redisServ},
		Failed:    []ExternalServer{},
		Connected: []ExternalServer{},
	}

	//connect all dependencies
	err := connManager.ConnectServices()

	if err != nil {
		connManager.PrintFailed()
		return
	}

	//start grpc server listener
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//create new grpc server and register its dependencies
	s := grpc.NewServer()
	pb.RegisterAuthServer(s, &resolver.Server{

		User: &db.UserModel{
			DB: sqlServ.conn,
		},
	})

	connManager.printConnected()

	log.Printf("Running grpc service on port: " + port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

//ConnectionManager manage connection timeouts
type ConnectionManager struct {
	Servers   []ExternalServer
	Failed    []ExternalServer
	Connected []ExternalServer
}

//PrintFailed prints failed dbs
func (c *ConnectionManager) printConnected() {

	for _, server := range c.Connected {
		log.Printf(server.Name() + " Connected Succefully")
	}
}

//PrintFailed prints failed dbs
func (c *ConnectionManager) PrintFailed() {

	for _, server := range c.Failed {
		println(server.Name() + " Failed")
	}
}

//ConnectServices attempts to connect all dbs
func (c *ConnectionManager) ConnectServices() error {

	for _, server := range c.Servers {

		err := server.AttemptConnect()

		if err != nil {
			c.Failed = append(c.Failed, server)
		} else {

			//add to connected list
			c.Connected = append(c.Connected, server)
		}
	}

	if len(c.Failed) > 0 {
		return fmt.Errorf("Some services failed to connect")
	}

	println(len(c.Failed))
	//no errors
	return nil
}

//ExternalServer interface
type ExternalServer interface {
	AttemptConnect() error
	Name() string
}

//RedisServer redis
type RedisServer struct {
	connected   bool
	host        string
	port        string
	dialTimeout time.Duration
	maxRetries  int
	conn        redis.Conn
}

//Name returns name of db
func (s *MySQLServer) Name() string {
	return "SQL Server"
}

//Name returns name of db
func (s *RedisServer) Name() string {
	return "Redis Server"
}

//AttemptConnect attempts to connect
func (s *RedisServer) AttemptConnect() error {

	currAttempts := 0

	for currAttempts < s.maxRetries && s.connected == false {

		println()

		conn, err := s.OpenRedis(s.host, s.port, s.dialTimeout)

		if err == nil {
			s.conn = conn
			s.connected = true
		} else {
			println(err.Error())
		}

		currAttempts++

	}

	if s.connected == false {
		return fmt.Errorf("Failed to connect")
	}

	return nil

}

//AttemptConnect attempts to connect
func (s *MySQLServer) AttemptConnect() error {

	currAttempts := 0

	for currAttempts < s.maxRetries && s.connected == false {

		conn, err := s.openDB(s.dsn)

		if err != nil {
			println(err.Error())
		} else {
			s.conn = conn
			s.connected = true
		}

		currAttempts++

	}

	if s.connected == false {
		return fmt.Errorf("Failed to connect")
	}

	return nil

}

//OpenRedis opens redis connection to server
func (s *RedisServer) OpenRedis(host, port string, timeout time.Duration) (redis.Conn, error) {
	conn, err := redis.Dial("tcp", host+":"+port, redis.DialConnectTimeout(timeout))

	if err != nil {
		println("Redis failed to connect at " + err.Error())
		return nil, err
	}

	return conn, nil
}

func (s *MySQLServer) openDB(dsn string) (*sql.DB, error) {

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		// print(err.Error())
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
