package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	firestore "cloud.google.com/go/firestore"
)

var app *firebase.App
var client *firestore.Client
func (h *Handler) setUpApp() {
	ProjectID := os.Getenv("ProjectID")
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: ProjectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}

func (h *Handler) authUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Test GET endpoint is being hit now!")
	ctx := context.Background()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("1")
		http.Error(w, err.Error(), 500)
		return
	}

	var idToken string
	err = json.Unmarshal(body, &idToken)
	if err != nil {
		log.Println("2")
		http.Error(w, err.Error(), 500)
		return
	}

	// auth, err := app.Auth(ctx)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// token, err := client.VerifyIDTokenAndCheckRevoked(ctx, idToken)
	// if err != nil {
	// 	log.Fatalf("error verifying ID token: %v\n", err)
	// }

	
	userInfo, err := client.Collection("users").Doc(idToken).Get(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	output, err := json.Marshal(userInfo.Data())
	if err != nil {
		log.Println("3")
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
	return
}

func (h *Handler) getColleges(w http.ResponseWriter, r *http.Request) {
	// ProjectID := os.Getenv("ProjectID")
	ctx := context.Background()
	log.Println("College endpoint")

	var colleges []college
	iter := client.Collection("Colleges").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
				break
		}
		if err != nil {
				return
		}
		fmt.Println(doc.Data())
		bs, err := json.Marshal(doc.Data())
		var tempCollege college
		err = json.Unmarshal(bs, &tempCollege)
		colleges = append(colleges, tempCollege)
	}
	output, err := json.Marshal(colleges)
	if err != nil {
		log.Println("3")
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
	return
}

func (h *Handler) getMatches(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var userUID string
	err = json.Unmarshal(body, &userUID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var user student
	dsnap, err := client.Collection("users").Doc(userUID).Get(ctx)
	dsnap.DataTo(&user)
	var colleges []college
	for _, c := range user.Matches {
		dataSnap, err := client.Collection("Colleges").Doc(c).Get(ctx)
		if err != nil {
			log.Fatalln(err)
		}
		bs, err := json.Marshal(dataSnap.Data())
		var tempCollege college
		err = json.Unmarshal(bs, &tempCollege)
		colleges = append(colleges, tempCollege)
	}
	output, err := json.Marshal(colleges)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
	return
}

type college struct {
	AcceptanceRate float64 `json:"Acceptance Rate"`
	AverageGPA	float64 `json:"Average GPA"`
	AverageSAT int64 `json:"Average SAT"`
	Diversity float32 `json:"Diversity"`
	Name string `json:"Name"`
	Size int64 `json:"Size"`
	Zip int64 `json:"Zip Code"`
}

type student struct {
	UID            string   `json:"uid"`
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	SchoolCode     string   `json:"schoolCode"`
	GraduationYear string   `json:"graduationYear"`
	WeightedGPA    float32  `json:"weightedGpa"`
	UnweightedGPA  float32  `json:"unweightedGpa"`
	ClassRank      int      `json:"classRank"`
	SAT            int      `json:"SAT"`
	ACT            int      `json:"ACT"`
	Size           string   `json:"size"`
	Location       string   `json:"location"`
	Diversity      string   `json:"diversity"`
	Majors         []string `json:"majors"`
	Distance       string   `json:"distance"`
	Zip            string   `json:"zip"`
	Matches        []string `json:"matches"`
}