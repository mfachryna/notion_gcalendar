package main

import (
	"context"
	"fmt"
	"log"
	"notioncalendar/src/domain"
	"notioncalendar/src/infra/driver"
	_taskRepository "notioncalendar/src/task/repository"
	"notioncalendar/src/util/env_driver"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/jomei/notionapi"
	"google.golang.org/api/calendar/v3"
)

func main() {

	env, err := env_driver.NewEnvDriver()
	if err != nil {
		log.Fatal(err.Error())
	}
	postgreConn, err := driver.NewPostgreConn(env.Postgre)
	if err != nil {
		log.Fatal(err.Error())
	}

	app := domain.App{
		PostgreDB: postgreConn,
	}
	taskRepository := _taskRepository.NewTaskRepository(app.PostgreDB)
	s := gocron.NewScheduler(time.Now().Location())
	calendarCon, _ := driver.NewGoogleConnection()

	s.Every(30).Day().StartImmediately().Do(func() {
		fmt.Println("---------------start calendar monthly api call---------------")
		year, month, day := time.Now().Date()
		timeMin := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		timeMax := time.Date(year, month, day+30, 23, 59, 59, 9999999, time.Local)
		GcalendarSyncMonth(taskRepository, calendarCon, timeMin.Format(time.RFC3339), timeMax.Format(time.RFC3339), "primary")
		GcalendarSyncMonth(taskRepository, calendarCon, timeMin.Format(time.RFC3339), timeMax.Format(time.RFC3339), "muhammad.fachry@suitmedia.com")
		fmt.Println("---------------end calendar monthly api call---------------")
		fmt.Println()
		fmt.Println()
	})

	s.Every(10).Second().Do(func() {
		fmt.Println("---------------start notion api call---------------")
		if err := callNotion(taskRepository); err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("---------------end notion api call---------------")
		fmt.Println()
		fmt.Println()
		fmt.Println("---------------start calendar api call---------------")
		tn := time.Now().Add(time.Minute * -6).Format(time.RFC3339)
		if err := callGcalendar(taskRepository, calendarCon, tn); err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("---------------end calendar api call---------------")
		fmt.Println()
		fmt.Println()
		time.Sleep(time.Second * 5)
	})
	s.StartBlocking()
}

func GcalendarSyncMonth(taskRepository domain.TaskRepository, calendarCon *calendar.Service, tn, tm, calendarId string) error {
	event, err := calendarCon.Events.List(calendarId).ShowDeleted(false).
		SingleEvents(true).TimeMin(tn).TimeMax(tm).MaxResults(200).OrderBy("startTime").Do()
	if err != nil {
		fmt.Println("Error while querying calendar", err.Error())
		return err
	}
	if err := taskRepository.InsertFromGoogleCalendar(event); err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func callGcalendar(taskRepository domain.TaskRepository, calendarCon *calendar.Service, tn string) error {
	event, err := calendarCon.Events.List("primary").ShowDeleted(true).
		SingleEvents(true).UpdatedMin(tn).MaxResults(5).OrderBy("startTime").Do()
	if err != nil {
		fmt.Println("Error while querying calendar", err.Error())
		return err
	}
	if err := taskRepository.InsertFromGoogleCalendar(event); err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
func callNotion(taskRepository domain.TaskRepository) error {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Load ENV error")
		return err
	}
	notionSecret := os.Getenv("NOTION_SECRET")
	client := notionapi.NewClient(notionapi.Token(notionSecret))
	tn := time.Now().Add(time.Minute * -6)
	res, err := client.Database.Query(context.Background(), notionapi.DatabaseID(os.Getenv("NOTION_DATABASE")), &notionapi.DatabaseQueryRequest{Filter: notionapi.PropertyFilter{Property: "Updated", Date: &notionapi.DateFilterCondition{OnOrAfter: (*notionapi.Date)(&tn)}}})
	if err != nil {
		fmt.Println("Fail querying DB")
		return err
	}
	if err := taskRepository.InsertFromNotion(res); err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
