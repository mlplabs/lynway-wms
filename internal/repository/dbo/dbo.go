package dbo

type UserData struct {
	Id       int
	Login    string
	Password string
	Code     string
}

type UpdateItemParams struct {
	ID             int
	Name           string
	ManufacturerID *int
	SKU            *string
	Barcode        *string
	Tags           []string
	Attributes     []byte
	UpdatedBy      *int
}

type CreateManufacturerParams struct {
	Name      string
	CreatedBy *int
}

type UpdateManufacturerParams struct {
	ID        int
	Name      string
	UpdatedBy *int
}

type CreateItemParams struct {
	Name           string
	ManufacturerID *int
	SKU            *string
	Barcode        *string
	Tags           []string
	Attributes     []byte
	CreatedBy      *int
}

type Printer struct {
	Id      int    // ID принтера в системе
	Name    string // Наименование принтера
	AgentId int    // ID агента, на котором настроен принтер
}
type CreateAgentInfoParams struct {
	Token     string
	Printers  []Printer
	CreatedBy *int
}

type UpdateAgentInfoParams struct {
	ID        int
	Token     string
	Printers  []Printer
	UpdatedBy *int
}

type User struct {
	ID       int
	Login    string
	Name     string
	IsActive bool
}

type CreateUserParams struct {
	Login     string
	Name      string
	IsActive  bool
	CreatedBy *int
}

type UpdateUserParams struct {
	ID        int
	Login     string
	Name      string
	IsActive  bool
	UpdatedBy *int
}
