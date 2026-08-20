package dto

type PrintTaskType int
type TaskStatus int

type EntityTypeId int

const (
	PrintTaskTypeLabel PrintTaskType = 1 // Стандартная этикетка

	PrintTaskStatusPending    TaskStatus = 0  // Ожидание
	PrintTaskStatusProcessing TaskStatus = 1  // В обработке
	PrintTaskStatusSuccess    TaskStatus = 11 // Завершилось успешно
	PrintTaskStatusError      TaskStatus = 12 // Завершилось с ошибкой

	EntityTypeIdProduct   EntityTypeId = 1 // Товары
	EntityTypeIdContainer EntityTypeId = 2 // Контейнеры
	EntityTypeIdTypeCell  EntityTypeId = 3 // Ячейки
)

type DecodeRequest struct {
	Code string `json:"code"` // Данные от сканера
}

type DecodeResponse struct {
	EntityTypeId int  `json:"type"` // Тип данных
	EntityId     *int `json:"id"`   // ID данных, если имеются
}

type GetPrintTaskRequest struct {
}

type GetPrintTaskResponse struct {
	Id           int               `json:"id"`             // ID задачи
	Type         PrintTaskType     `json:"type"`           // Тип задачи
	PrinterName  string            `json:"printer_name" `  // Название принтера, на который надо печатать
	EntityId     int               `json:"entity_id"`      // ID данных
	EntityTypeId EntityTypeId      `json:"entity_type_id"` // ID типа данных
	Data         string            `json:"data"`           // RAW (base64) данные для печати
	Labels       map[string]string `json:"labels"`		   // Типовые метки
	ToFile       bool              `json:"to_file"` 	   // Печать в файл (в разработке)
}

// AgentInfoRequest информация об агенте
type AgentInfoRequest struct {
	Name     string        `json:"name"`     // Название агента (в разработке)
	Version  string        `json:"version"`  // Версия агента (в разработке)
	Printers []PrinterInfo `json:"printers"` // Список принтеров
}

// PrinterInfo информация о принтере
type PrinterInfo struct {
	PrinterName string `json:"printer_name"` // Название принтера
}

// UpdatePrintTaskStatusRequest - запрос на обновление задачи на печать
type UpdatePrintTaskStatusRequest struct {
	Id     int        `json:"id"`
	Status TaskStatus `json:"status"`
}
