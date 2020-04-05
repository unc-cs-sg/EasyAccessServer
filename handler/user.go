package handler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	firestore "cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"google.golang.org/api/iterator"
)

//Verify user
func Verify(idToken string) (*auth.Token, error) {
	ctx := context.Background()
	temp, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	token, err := temp.VerifyIDTokenAndCheckRevoked(ctx, idToken)
	if err != nil {
		return nil, err
	}
	return token, nil
}

//AuthUser is
func (h *Handler) AuthUser(w http.ResponseWriter, r *http.Request) {
	log.Println("USER")
	ctx := context.Background()
	idToken := r.Header.Get("Authorization")
	token, err := Verify(idToken)
	if err != nil {
		log.Println("error verifying ID token: ", err.Error())
		http.Error(w, err.Error(), 401)
		return
	}

	userInfo, err := client.Collection("users").Doc(token.UID).Get(ctx)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 404)
		return
	}

	output, err := json.Marshal(userInfo.Data())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
	return
}

func (h *Handler) addUserInfo(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	idToken := r.Header.Get("Authorization")
	token, err := Verify(idToken)
	if err != nil {
		log.Println("error verifying ID token: ", err.Error())
		http.Error(w, err.Error(), 401)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	var user student
	err = json.Unmarshal(body, &user)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = client.Collection("users").Doc(token.UID).Set(ctx, user)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

type updateInfo struct {
	UID  string             `json:"uid"`
	Info []firestore.Update `json:"info"`
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	idToken := r.Header.Get("Authorization")
	token, err := Verify(idToken)
	if err != nil {
		log.Println("error verifying ID token: ", err.Error())
		http.Error(w, err.Error(), 401)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var newInfo *updateInfo
	err = json.Unmarshal(body, &newInfo)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	userRef := client.Collection("users").Doc(token.UID)
	_, err = userRef.Update(ctx, newInfo.Info)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return

}

func scoreStudent(UID string) (int, string, error) {
	ctx := context.Background()

	userInfo, err := client.Collection("users").Doc(UID).Get(ctx)
	if err != nil {
		log.Println(err)
	}

	var student student
	userInfo.DataTo(&student)

	var potentialScores []string
	iter := client.Collection("Selectivity").Where("LowGPA", "<=", student.UnweightedGPA).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, "", err
		}
		var s selectivity
		doc.DataTo(&s)
		if s.HighACT != 0 {
			if s.LowACT <= student.ACT && student.ACT <= s.HighACT && s.LowGPA <= float64(student.UnweightedGPA) && float64(student.UnweightedGPA) <= s.HighGPA {
				potentialScores = append(potentialScores, s.Score)
			}
		} else {
			if s.LowSAT <= student.SAT && student.SAT <= s.HighSAT && s.LowGPA <= float64(student.UnweightedGPA) && float64(student.UnweightedGPA) <= s.HighGPA {
				potentialScores = append(potentialScores, s.Score)
			}
		}
	}
	topScore := 0
	topTest := ""
	for _, score := range potentialScores {
		i, err := strconv.Atoi(score[len(score)-1:])
		if err != nil {
			return 0, "", err
		}
		if i > topScore {
			topScore = i
			topTest = score[len(score)-4 : len(score)-1]
		}
	}
	if topScore == 0 {
		topScore = 1
	}
	return topScore, topTest, nil
}

type student struct {
	UID           string  `json:"uid"`
	FirstName     string  `json:"firstname"`
	LastName      string  `json:"lastname"`
	Email         string  `json:"email"`
	SchoolCode    string  `json:"schoolCode"`
	SchoolName    string  `json:"schoolName"`
	WeightedGPA   float32 `json:"weightedGpa"`
	UnweightedGPA float32 `json:"unweightedGpa"`
	ClassRank     int     `json:"classRank"`
	SAT           int     `json:"SAT"`
	ACT           int     `json:"ACT"`
	State         string  `json:"state"`
	Zip           string  `json:"zip"`
}
