package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/criyle/go-judge/pb"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

var upgrader = websocket.Upgrader{}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 50 * time.Second
)

func main() {
	// connect to database
	db := getDB()

	grpcAddr := "localhost:5051"
	if s := os.Getenv("GRPC_SERVER"); s != "" {
		grpcAddr = s
	}

	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalln("client", err)
	}
	client := pb.NewExecutorClient(conn)

	// creates client hub
	ch := newClientHub()
	go ch.loop()
	http.HandleFunc("/ws", ch.handleWS)

	// creates judger hub
	jh := newJudgerHub(db, ch)
	go jh.loop()
	http.HandleFunc("/jws", jh.handleJudgerWS)

	// creates shell handle
	sh := &shell{db: db, client: client}
	http.HandleFunc("/shell", sh.handleShellWS)

	// REST api
	a := api{db: db, judgerHub: jh}
	http.Handle("/api/", http.StripPrefix("/api", a.serveMux()))

	// static files for spa
	s := spaFS{FileSystem: http.Dir("dist")}
	http.Handle("/", http.FileServer(s))

	// local test env
	addr := ":5000"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	// start serve
	log.Fatal(http.ListenAndServe(addr, nil))
}

type spaFS struct {
	http.FileSystem
}

func (fs spaFS) Open(name string) (http.File, error) {
	f, err := fs.FileSystem.Open(name)
	if err != nil && os.IsNotExist(err) {
		return fs.FileSystem.Open("index.html")
	}
	return f, err
}
