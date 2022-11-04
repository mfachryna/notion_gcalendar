package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"notioncalendar/src/domain"
	"notioncalendar/src/infra/driver"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/joho/godotenv"
	"github.com/jomei/notionapi"
	"google.golang.org/api/calendar/v3"
	"gorm.io/gorm"
)

type TaskRepository struct {
	DB *gorm.DB
}

func NewTaskRepository(db *gorm.DB) domain.TaskRepository {
	return &TaskRepository{
		DB: db,
	}
}

func (t *TaskRepository) InsertFromNotion(res *notionapi.DatabaseQueryResponse) error {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Load ENV error")
		return err
	}
	if len(res.Results) > 0 {
		for _, item := range res.Results {
			taskDb, calendarEvent, err := createNotionObject(item)
			if err != nil || taskDb == nil || calendarEvent == nil {
				fmt.Println("Initialize data fail")
				return err
			}
			var taskFind domain.Task
			t.DB.Where("id", item.ID).First(&taskFind)

			calendarCon, err := driver.NewGoogleConnection()
			if err != nil {
				return err
			}
			email := os.Getenv("CALENDAR_EMAIL")
			email2 := os.Getenv("CALENDAR_EMAIL2")
			if taskDb.Creator != email && taskDb.Creator != email2 {
				fmt.Println("Can't send because of user email")
				continue
			}
			calendarId := "primary"
			if taskDb.Creator == email2 {
				calendarId = email2
			}
			tx := t.DB.Begin()
			if taskFind.Id == "" {
				if !(taskDb.AddCalendar && taskDb.PlannedStart != nil) {
					fmt.Println("Continuing notion because not planned")
					continue
				}
				if taskDb.Status == "Canceled" || taskDb.Status == "Hold" {
					calendarEvent.Status = "cancelled"
				}
				dbc := tx.Create(&taskDb)

				if dbc.Error != nil {
					fmt.Println("Creating data error", dbc.Error.Error())
					tx.Rollback()
					continue
				}
				calendarEvent.Reminders = &calendar.EventReminders{UseDefault: true}
				evenCon, err := calendarCon.Events.Insert(calendarId, calendarEvent).Do()
				if err != nil {
					fmt.Println("Error inserting to calendar", err.Error())
					continue
				}

				oldDbTask := domain.Task{}
				copier.Copy(&oldDbTask, taskDb)
				taskDb.CalendarID = evenCon.Id
				taskDb.CalendarUrl = evenCon.HangoutLink
				dbu := tx.Model(&oldDbTask).Where("id = ?", taskDb.Id).Updates(&taskDb)
				if dbu.Error != nil {
					fmt.Println("Updating clendar data from notion error", dbu.Error.Error())
					tx.Rollback()
					continue
				}
				fmt.Println("Success creating data from notion")
			} else {
				if !(taskDb.AddCalendar && taskDb.PlannedStart != nil) {

					dba := tx.Model(&taskFind).Where("id = ?", taskFind.Id).Delete(&taskDb)
					if dba.Error != nil {
						fmt.Println(dba.Error.Error())
						continue
					}
					err := calendarCon.Events.Delete(calendarId, taskFind.CalendarID).Do()
					if err != nil {
						fmt.Println("Error deleting calendar", err.Error())
						tx.Rollback()
						continue
					}
					fmt.Println("Continuing notion because not planned")
					continue
				}
				if taskDb.Status == "Canceled" || taskDb.Status == "Hold" {
					err := calendarCon.Events.Delete(calendarId, taskFind.CalendarID).Do()
					if err != nil {
						fmt.Println("Error deleting calendar", err.Error())
						tx.Rollback()
						continue
					}
				} else {
					calendarEvent.Status = "confirmed"
				}
				dba := tx.Model(&taskFind).Where("id = ?", taskFind.Id).Updates(&taskDb)
				if dba.Error != nil {
					if dba.Error.Error() == "Nothing changed" {
						fmt.Println(dba.Error.Error())
						continue
					}
					fmt.Println("Updating clendar data from notion error")
					tx.Rollback()
					return dba.Error
				}

				_, err := calendarCon.Events.Update(calendarId, taskFind.CalendarID, calendarEvent).Do()
				if err != nil {
					fmt.Println("Error updating to calendar", err.Error())
					tx.Rollback()
					continue
				}
				fmt.Println("Success updating data from notion")
			}
			tx.Commit()
			time.Sleep(time.Second * 1)
		}
	}

	return nil
}

func (t *TaskRepository) InsertFromGoogleCalendar(event *calendar.Events) error {

	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Load ENV error")
		return err
	}
	notionSecret := os.Getenv("NOTION_SECRET")
	client := notionapi.NewClient(notionapi.Token(notionSecret))
	for _, item := range event.Items {

		dbTask, taskProperties := CreateCalendarObject(item)

		unprocessTask := notionapi.PageCreateRequest{
			Parent: notionapi.Parent{
				Type:       "database_id",
				DatabaseID: notionapi.DatabaseID(os.Getenv("NOTION_DATABASE")),
			},
			Properties: taskProperties,
		}
		var taskFind domain.Task
		t.DB.Where("calendar_id", dbTask.CalendarID).First(&taskFind)
		tx := t.DB.Begin()
		if taskFind.CalendarID == "" {
			if item.Status == "cancelled" {
				fmt.Println("Not creating event because data has been deleted")
				continue
			}
			status := notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: "To Do",
				},
			}

			taskProperties["Status"] = status
			dbTask.Status = "To Do"
			if dbTask.Due == nil && dbTask.PlannedEnd != nil {
				due := notionapi.DateProperty{
					Date: &notionapi.DateObject{
						Start: (*notionapi.Date)(dbTask.PlannedEnd),
					},
				}
				taskProperties["Due"] = due
			} else if dbTask.Due == nil && dbTask.PlannedStart != nil {
				due := notionapi.DateProperty{
					Date: &notionapi.DateObject{
						Start: (*notionapi.Date)(dbTask.PlannedStart),
					},
				}
				taskProperties["Due"] = due
			}

			dbc := tx.Create(&dbTask)
			if dbc.Error != nil {
				fmt.Println("Creating data error", dbc.Error.Error())
				tx.Rollback()
				continue
			}
			page, err := client.Page.Create(context.Background(), &unprocessTask)
			if err != nil {
				fmt.Println("Fail creating notion page", err.Error())
				tx.Rollback()
				continue
			}
			oldDbTask := domain.Task{}
			copier.Copy(&oldDbTask, dbTask)
			dbTask.NotionUrl = page.URL
			dbTask.Id = page.ID.String()
			dba := tx.Model(&oldDbTask).Where("calendar_id = ?", dbTask.CalendarID).Updates(&dbTask)
			if dba.Error != nil {
				fmt.Println("Updating notion id error", dba.Error.Error())
				tx.Rollback()
				continue
			}
			fmt.Println("Success creating data from calendar")
		} else {
			dba := tx.Model(&taskFind).Where("calendar_id = ?", taskFind.CalendarID).Updates(&dbTask)
			if dba.Error != nil {
				if dba.Error.Error() == "Nothing changed" {
					fmt.Println(dba.Error.Error())
					continue
				}
				fmt.Println("Updating clendar data from notion error")
				tx.Rollback()
				return dba.Error
			}
			if item.Status == "cancelled" && dbTask.Status == "" {
				fmt.Println("Not creating event because data has been deleted")
				taskProperties["Status"] = notionapi.SelectProperty{
					Select: notionapi.Option{
						Name: "Canceled",
					},
				}
			}
			_, err := client.Page.Update(context.Background(), notionapi.PageID(taskFind.Id), &notionapi.PageUpdateRequest{Properties: unprocessTask.Properties})
			if err != nil {
				fmt.Println("Fail updating notion page", err.Error())
				tx.Rollback()
				continue
			}

			fmt.Println("Success updating data from calendar")
		}
		tx.Commit()
		time.Sleep(time.Second * 1)
	}
	return nil
}

func CreateCalendarObject(item *calendar.Event) (*domain.Task, notionapi.Properties) {
	var dateStart *time.Time
	if item.Start.DateTime == "" {
		fmt.Println("date")
		dateTemp, err := time.Parse("2001-01-01", item.Start.Date)
		if err != nil {
			dateStart = nil
		} else {
			dateStart = &dateTemp
		}
	} else {
		dateTemp, err := time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			fmt.Println(err.Error())
			dateStart = nil
		} else {
			dateStart = &dateTemp
		}
	}

	var dateEnd *time.Time
	if item.End.DateTime == "" {
		dateTemp, err := time.Parse("2001-01-01", item.End.Date)
		if err != nil {
			fmt.Println(err.Error())
			dateEnd = nil
		} else {
			dateEnd = &dateTemp
		}
	} else {
		dateTemp, err := time.Parse(time.RFC3339, item.End.DateTime)
		if err != nil {
			fmt.Println(err.Error())
			dateEnd = nil
		} else {
			dateEnd = &dateTemp
		}
	}
	Attendees := ""
	if len(item.Attendees) >= 0 {
		for _, itemAttend := range item.Attendees {
			Attendees = Attendees + itemAttend.Email + ";"
		}
	}
	dbTask := &domain.Task{
		CalendarID:   item.Id,
		Name:         item.Summary,
		AddCalendar:  true,
		Archive:      false,
		Assignee:     Attendees,
		Creator:      item.Creator.Email,
		PlannedStart: dateStart,
		PlannedEnd:   dateEnd,
		MeetingLink:  item.HangoutLink,
		CalendarUrl:  item.HtmlLink,
	}
	properties := make(notionapi.Properties)
	tasks := notionapi.TitleProperty{
		Title: []notionapi.RichText{
			{
				Text: &notionapi.Text{
					Content: dbTask.Name,
				},
			},
		},
	}
	properties["Task"] = tasks
	properties["Creator"] = notionapi.EmailProperty{
		Email: dbTask.Creator,
	}

	if dbTask.Priority != "" {
		priority := notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: dbTask.Priority,
			},
		}
		properties["Priority"] = priority
	}
	if dbTask.PlannedStart != nil {
		plannedDate := notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: (*notionapi.Date)(dbTask.PlannedStart),
				End:   (*notionapi.Date)(dbTask.PlannedEnd),
			},
		}
		properties["Planned Date"] = plannedDate
	}
	overview := notionapi.RichTextProperty{
		RichText: []notionapi.RichText{
			{
				Text: &notionapi.Text{
					Content: "",
				},
			},
		},
	}
	properties["Overview"] = overview
	if dbTask.Type != "" {
		types := notionapi.SelectProperty{
			Select: notionapi.Option{
				Name: dbTask.Type,
			},
		}
		properties["Type"] = types
	}
	archive := notionapi.CheckboxProperty{
		Checkbox: dbTask.Archive,
	}
	properties["Archive"] = archive
	addCalendar := notionapi.CheckboxProperty{
		Checkbox: dbTask.AddCalendar,
	}
	properties["Add Calendar"] = addCalendar
	if dbTask.Due != nil {
		due := notionapi.DateProperty{
			Date: &notionapi.DateObject{
				Start: (*notionapi.Date)(dbTask.Due),
			},
		}
		properties["Due"] = due
	}
	if dbTask.MeetingLink != "" {
		meetingLink := notionapi.URLProperty{
			Type: "url",
			URL:  dbTask.MeetingLink,
		}
		properties["Meeting Link"] = meetingLink
	}
	assignee := notionapi.RichTextProperty{
		RichText: []notionapi.RichText{
			{
				Text: &notionapi.Text{
					Content: dbTask.Assignee,
				},
			},
		},
	}
	properties["Assignee"] = assignee
	return dbTask, properties
}

func createNotionObject(item notionapi.Page) (*domain.Task, *calendar.Event, error) {
	var calendarEvent calendar.Event

	resByte, err := json.Marshal(item.Properties)
	if err != nil {
		fmt.Println(err.Error())
		return nil, nil, err
	}
	var tasks domain.NotionTask
	json.Unmarshal(resByte, &tasks)
	//insert to db
	var taskDb domain.Task
	taskDb.Id = item.ID.String()
	if len(tasks.Task.Title) > 0 {
		taskDb.Name = tasks.Task.Title[0].PlainText
		calendarEvent.Summary = tasks.Task.Title[0].PlainText
	}
	taskDb.AddCalendar = tasks.AddCalendar.Checkbox
	taskDb.Archive = tasks.Archive.Checkbox
	if len(tasks.Assignee.RichText) > 0 {
		taskDb.Assignee = tasks.Assignee.RichText[0].PlainText
		s := strings.Split(tasks.Assignee.RichText[0].PlainText, ";")
		for _, assign := range s {
			if assign != "" {
				calendarEvent.Attendees = append(calendarEvent.Attendees, &calendar.EventAttendee{Email: assign})
			}
		}
	}
	if len(tasks.Overview.RichText) > 0 {
		taskDb.Assignee = tasks.Overview.RichText[0].PlainText
		calendarEvent.Description = tasks.Overview.RichText[0].Text.Content
	}
	if (tasks.Due.Date) != nil {
		taskDb.Due = (*time.Time)(tasks.Due.Date.End)
	}

	taskDb.Status = tasks.Status.Select.Name
	if (tasks.PlannedDate.Date) != nil {
		taskDb.PlannedStart = (*time.Time)(tasks.PlannedDate.Date.Start)
		calendarEvent.Start = &calendar.EventDateTime{DateTime: ((*time.Time)(tasks.PlannedDate.Date.Start)).Format(time.RFC3339)}
		fmt.Println("test")
		if (tasks.PlannedDate.Date.End) != nil {
			calendarEvent.EndTimeUnspecified = false
			calendarEvent.End = &calendar.EventDateTime{DateTime: ((*time.Time)(tasks.PlannedDate.Date.End)).Format(time.RFC3339)}
			taskDb.PlannedEnd = (*time.Time)(tasks.PlannedDate.Date.End)
		} else {

			year, month, day := (*time.Time)(tasks.PlannedDate.Date.Start).Date()
			calendarEvent.Start = &calendar.EventDateTime{DateTime: time.Date(year, month, day, 0, 0, 0, 0, time.Local).Format(time.RFC3339)}
			calendarEvent.End = &calendar.EventDateTime{DateTime: time.Date(year, month, day, 23, 59, 59, 9999999, time.Local).Format(time.RFC3339)}
		}
	}
	taskDb.CreatedAt = tasks.Created.CreatedTime
	taskDb.CreatedAt = item.LastEditedTime
	taskDb.MeetingLink = tasks.MeetingLink.URL
	taskDb.Type = tasks.Type.Select.Name
	color := domain.GetTaskTypeColor(taskDb.Type)
	if color != "" {
		calendarEvent.ColorId = color
	}
	if len(tasks.Context.MultiSelect) > 0 {
		for _, context := range tasks.Context.MultiSelect {
			taskDb.Context += context.Name
		}
	}
	taskDb.NotionUrl = item.URL
	taskDb.Creator = tasks.Creator.Email

	return &taskDb, &calendarEvent, nil
}
