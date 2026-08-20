package model

import "github.com/mlplabs/lynway-wms/pkg/dto"

// Agent внешний воркер, находящийся на локальной машине
// получающий задание с сервера
type Agent struct {
	Id    int
	Name  *string
	Token *string
}

// Printer принтер для печати
type Printer struct {
	Id      int    // ID принтера в системе
	Name    string // Наименование принтера
	AgentId int    // ID агента, на котором настроен принтер
}

// PrintQueueTask элемент очереди печати
type PrintQueueTask struct {
	Id           int               // ID задачи
	PrintType    dto.PrintTaskType // Тип печати
	Printer      *Printer          // Принтер + агент
	EntityId     int               // ID данных
	EntityTypeId dto.EntityTypeId  // Тип данных
}

type PrintTask struct {
	Id           int               // ID задачи
	Type         dto.PrintTaskType // Тип задачи
	PrinterName  string            // Название принтера, на который надо печатать
	EntityId     int               // ID данных
	EntityTypeId dto.EntityTypeId  // ID типа данных
	Data         string            // RAW (base64) данные для печати
	Labels       map[string]string // Доп информация для этикетки
	ToFile       bool              // Признак печати в файл
}

type AgentInfo struct {
	Token    string
	Printers []Printer
}

func (a *AgentInfo) FromDTO(ar *dto.AgentInfoRequest) *AgentInfo {
	for _, p := range ar.Printers {
		a.Printers = append(a.Printers, Printer{
			Name: p.PrinterName,
		})
	}
	return a
}

type PrintTaskStatus struct {
	Id     int
	Status dto.TaskStatus
}

func (pts *PrintTaskStatus) FromDTO(ptsr *dto.UpdatePrintTaskStatusRequest) *PrintTaskStatus {
	pts.Id = ptsr.Id
	pts.Status = ptsr.Status
	return pts
}
