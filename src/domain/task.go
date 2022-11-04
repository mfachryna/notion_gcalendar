package domain

import (
	"fmt"
	"time"

	"github.com/jomei/notionapi"
	"google.golang.org/api/calendar/v3"
	"gorm.io/gorm"
)

type Task struct {
	Id           string `gorm:"primaryKey"`
	Name         string
	Priority     string
	AddCalendar  bool
	Archive      bool
	Assignee     string
	Due          *time.Time `sql:"type:timestamp without time zone"`
	PlannedStart *time.Time `sql:"type:timestamp without time zone"`
	PlannedEnd   *time.Time `sql:"type:timestamp without time zone"`
	CreatedAt    time.Time  `sql:"type:timestamp without time zone"`
	UpdatedAt    time.Time  `sql:"type:timestamp without time zone"`
	MeetingLink  string
	Type         string
	Status       string
	Context      string
	CalendarID   string
	CalendarUrl  string
	Creator      string
	NotionUrl    string
	Overview     string
}
type TaskRepository interface {
	InsertFromNotion(*notionapi.DatabaseQueryResponse) error
	InsertFromGoogleCalendar(*calendar.Events) error
}

func GetTaskTypeColor(taskType string) string {
	colorMap := map[string]string{
		"Event":    "8",
		"Meeting":  "2",
		"Personal": "11",
		"Reminder": "5",
		"Task":     "3",
	}
	if val, ok := colorMap[taskType]; ok {
		return val
	}
	return ""
}

func (t *Task) BeforeUpdate(tx *gorm.DB) (err error) {
	if !(tx.Statement.Changed("Id") || tx.Statement.Changed("Name") || tx.Statement.Changed("Priority") || tx.Statement.Changed("AddCalendar") || tx.Statement.Changed("Archive") || tx.Statement.Changed("Assignee") || tx.Statement.Changed("Due") || tx.Statement.Changed("PlannedStart") || tx.Statement.Changed("PlannedEnd") || tx.Statement.Changed("MeetingLink") || tx.Statement.Changed("Type") || tx.Statement.Changed("Context") || tx.Statement.Changed("CalendarID") || tx.Statement.Changed("CalendarUrl") || tx.Statement.Changed("Creator") || tx.Statement.Changed("NotionUrl") || tx.Statement.Changed("Overview") || tx.Statement.Changed("Status")) {
		return fmt.Errorf("Nothing changed")
	}
	return
}
