package main

// Add event using Google Calendar API.
// https://github.com/google/google-api-go-client/blob/master/calendar/v3/calendar-gen.go

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ogiekako/pmdr/calendarc"
	"golang.org/x/net/context"
	"google.golang.org/api/calendar/v3"
)

var after = flag.Duration("after", 0, "Start first pomodoro after this duration.")
var from = flag.String("from", "", "Start time in the form of 15:04.")
var remove = flag.Bool("remove", false, "Remove the entries after the specified time")

var pmdrDuration = 25 * time.Minute
var shortBreak = 5 * time.Minute
var longBreak = 15 * time.Minute
var chunk = 4

func format(t time.Time) string {
	return t.Format(time.RFC3339)
}

func startTime() time.Time {
	t := time.Now().Add(*after)
	if *from != "" {
		year, month, day := t.Date()
		t2, err := time.Parse("15:04", *from)
		if err != nil {
			log.Fatalf("Failed to parse %s: %v\n", &from, err)
		}
		hour, min, _ := t2.Clock()
		t = time.Date(year, month, day, hour, min, 0, 0, t.Location())
	}
	return t
}

func removeEvents(srv *calendar.Service) {
	t := startTime()
	events, err := srv.Events.List("primary").ShowDeleted(false).SingleEvents(true).TimeMin(format(t)).TimeMax(format(t.Add(24 * time.Hour))).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve events. %v", err)
	}
	for _, item := range events.Items {
		if strings.Contains(item.Summary, "pmdr") {
			fmt.Printf("Removing %s\n", item.Summary)
			if err := srv.Events.Delete("primary", item.Id).Do(); err != nil {
				log.Fatalf("Failed to delete event. %v", err)
			}
		}
	}
}

func createEvent(srv *calendar.Service, fromId, count int) {
	t := startTime()

	for i := fromId; i < fromId+count; i++ {
		summary := fmt.Sprintf("pmdr %d", i)
		var remindBefore time.Duration
		if (i-fromId)%chunk == 0 {
			remindBefore = longBreak
		} else {
			remindBefore = shortBreak
		}
		event := &calendar.Event{
			Summary:     summary,
			Description: "pomodoro",
			Start: &calendar.EventDateTime{
				DateTime: format(t),
				TimeZone: "Asia/Tokyo",
			},
			End: &calendar.EventDateTime{
				DateTime: format(t.Add(pmdrDuration)),
				TimeZone: "Asia/Tokyo",
			},
			Visibility: "private",
			Reminders: &calendar.EventReminders{
				Overrides: []*calendar.EventReminder{
					&calendar.EventReminder{
						Method:  "popup",
						Minutes: int64(remindBefore / time.Minute),
					},
					&calendar.EventReminder{
						Method:          "popup",
						Minutes:         0,
						ForceSendFields: []string{"Minutes"},
					},
				},
				UseDefault:      false,
				ForceSendFields: []string{"UseDefault"},
			},
		}
		calendarId := "primary"
		event, err := srv.Events.Insert(calendarId, event).Do()
		if err != nil {
			log.Fatalf("Unable to create event. %v\n", err)
		}
		fmt.Printf("%v-%v %v %s\n", t.Format("15:04"), t.Add(pmdrDuration).Format("15:04"), summary, event.HtmlLink)

		add := pmdrDuration
		if (i-fromId)%chunk == chunk-1 {
			add += longBreak
		} else {
			add += shortBreak
		}
		t = t.Add(add)
	}
}

func nextPmdrId(srv *calendar.Service) int {
	t := time.Now()
	year, month, day := t.Date()
	// Set 04:00 as the beginning of a day.
	origin := time.Date(year, month, day, 4, 0, 0, 0, t.Location())
	if origin.After(t) {
		origin = origin.Add(-24 * time.Hour)
	}
	events, err := srv.Events.List("primary").ShowDeleted(false).SingleEvents(true).TimeMin(format(origin)).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve today's user events. You may want to remove ~/.credentials/calendar-go-quickstart.json: %v", err)
	}
	res := 1
	for _, item := range events.Items {
		s := item.Summary
		if strings.Contains(s, "pmdr") {
			i, _ := strconv.Atoi(strings.TrimPrefix(s, "pmdr "))
			if res < i+1 {
				res = i + 1
			}
		}
	}
	return res
}

func main() {
	flag.Parse()
	count := 4

	if flag.NArg() == 0 {
		// Do nothing.
	} else if flag.NArg() == 1 {
		count, _ = strconv.Atoi(flag.Arg(0))
	} else {
		fmt.Fprintf(os.Stderr, "add [-after duration] [-from start_time] [count]")
		return
	}

	ctx := context.Background()
	srv := calendarc.NewService(ctx)

	if *remove {
		removeEvents(srv)
	} else {
		from := nextPmdrId(srv)
		createEvent(srv, from, count)
	}
}
