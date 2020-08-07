package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/chime"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js/"))))
	http.HandleFunc("/", joinMeeting)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("started server localhost:%s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func joinMeeting(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))

	sess := session.Must(session.NewSession())

	svc := chime.New(sess, aws.NewConfig().WithRegion("ap-northeast-1"))
	result, err := svc.ListMeetings(nil)
	if err != nil {
		panic(err)
	}

	// meetingがなければ新規作成しておく
	var mtg *chime.Meeting
	if len(result.Meetings) == 0 {
		res, err := svc.CreateMeeting(&chime.CreateMeetingInput{
			ClientRequestToken: aws.String(strconv.Itoa(int(time.Now().Unix()))),
			MediaRegion:        aws.String("ap-northeast-1"),
		})
		if err != nil {
			panic(err)
		}
		mtg = res.Meeting
	} else {
		mtg = result.Meetings[0]
	}

	// meetingにattendeeを追加
	atd, err := svc.CreateAttendee(&chime.CreateAttendeeInput{
		MeetingId:      mtg.MeetingId,
		ExternalUserId: aws.String(strconv.Itoa(int(time.Now().Unix()))),
	})
	if err != nil {
		panic(err)
	}

	// 確認用にidを出す
	out := struct {
		Meeting  string
		Attendee string
	}{
		Meeting:  string(jsonMtg(mtg)),
		Attendee: string(jsonAtd(atd.Attendee)),
	}
	log.Println(out)

	tmpl.Execute(w, out)
}

func jsonMtg(mtg *chime.Meeting) []byte {
	res, err := json.Marshal(mtg)
	if err != nil {
		panic(err)
	}
	return res
}

func jsonAtd(atd *chime.Attendee) []byte {
	res, err := json.Marshal(atd)
	if err != nil {
		panic(err)
	}
	return res
}
