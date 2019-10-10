package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

  firebase "firebase.google.com/go"
	"github.com/BaileyFrederick/EasyAccessServer/handler"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
  "github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	// loads values from .env into the system
	if err := godotenv.Load(".env"); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
  ProjectID := os.Getenv("ProjectID")
  
  
  serviceAccount := os.Getenv("SERVICE_ACCOUNT_ID")
  
	println("GOPATH set up correctly amd project is working")
	ctx := context.Background()

	conf := &firebase.Config{ProjectID: ProjectID, ServiceAccountID: serviceAccount}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	auth, err := app.Auth(ctx)
	userRecord, err := auth.GetUserByEmail(ctx, "FrederickBailey18@gmail.com")

	println(userRecord.UID)

	//test to change info in firestore
	p := user{
		Name: "Our app is hosted MF",
	}
	//Changes the name of the specific user based on UID to ALICE
	_, err = client.Collection("users").Doc("755O4T422rS1CgngVpI8").Set(ctx, p)
	if err != nil {
		log.Fatal(err)
	}

	err = setHandler()
	if err != nil {
		log.Println(err)
	}
}

func setHandler() error {
	// set up our global handler
	h, err := handler.New(handler.Config{
		Logger: log,
	})
	if err != nil {
		return errors.Wrap(err, "handler new")
	}

	println(h)
  port := os.Getenv("PORT")
  println(port)
	server := &http.Server{
		Handler: h,
		Addr:    fmt.Sprintf(":%v", port),
	}

	// do graceful server shutdown
	go gracefulShutdown(server, time.Second*30)
  
	log.Infof("listening on port %v", port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return errors.Wrap(err, "cannot start a server")
	}
	return nil
}

type user struct {
	Name string
}

// gracefulShutdown shuts down our server in a graceful way.
func gracefulShutdown(server *http.Server, timeout time.Duration) {
	sigStop := make(chan os.Signal)

	signal.Notify(sigStop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	<-sigStop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("graceful shutdown failed")
	}

	log.Info("graceful shutdown complete")
}
