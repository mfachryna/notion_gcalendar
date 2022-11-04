package domain

import "github.com/jomei/notionapi"

type NotionTask struct {
	ParentDueDate  notionapi.RollupProperty      `json:"Parent Due Date"`
	Priority       notionapi.SelectProperty      `json:"Priority"`
	PlannedDate    notionapi.DateProperty        `json:"Planned Date"`
	Overview       notionapi.RichTextProperty    `json:"Overview"`
	Creator        notionapi.EmailProperty       `json:"Creator"`
	Created        notionapi.CreatedTimeProperty `json:"Created"`
	State          notionapi.FormulaProperty     `json:"State"`
	QuickCapture   notionapi.FormulaProperty     `json:"Quick Capture"`
	CreatedBy      notionapi.CreatedByProperty   `json:"Created By"`
	Hold           notionapi.FormulaProperty     `json:"Hold"`
	Type           notionapi.SelectProperty      `json:"Type"`
	RealDue        notionapi.FormulaProperty     `json:"Real Due"`
	Canceled       notionapi.FormulaProperty     `json:"Canceled"`
	ParentProject  notionapi.RollupProperty      `json:"Parent Project"`
	Archive        notionapi.CheckboxProperty    `json:"Archive"`
	Cold           notionapi.FormulaProperty     `json:"Cold"`
	AddCalendar    notionapi.CheckboxProperty    `json:"Add Calendar"`
	Done           notionapi.FormulaProperty     `json:"Doner"`
	IsRecure       notionapi.FormulaProperty     `json:"Is recure"`
	Project        notionapi.RelationProperty    `json:"Project"`
	Due            notionapi.DateProperty        `json:"Due"`
	Notes          notionapi.RelationProperty    `json:"Notes"`
	MeetingLink    notionapi.URLProperty         `json:"Meeting Link"`
	RecureInterval notionapi.NumberProperty      `json:"Recure Interval"`
	Assignee       notionapi.RichTextProperty    `json:"Assignee"`
	RecureUnit     notionapi.SelectProperty      `json:"Recure Unit"`
	SubSeed        notionapi.FormulaProperty     `json:"Sub Seed"`
	ParentTask     notionapi.RelationProperty    `json:"Parent Task"`
	SubTask        notionapi.RelationProperty    `json:"Sub Task"`
	SubSeedName    notionapi.FormulaProperty     `json:"Sub Seed Name"`
	Status         notionapi.SelectProperty      `json:"Status"`
	Late           notionapi.FormulaProperty     `json:"Late"`
	Context        notionapi.MultiSelectProperty `json:"Context"`
	Task           notionapi.TitleProperty       `json:"Task"`
}
