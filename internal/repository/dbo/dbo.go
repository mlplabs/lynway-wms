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
}

type CreateItemParams struct {
	Name           string
	ManufacturerID *int
	SKU            *string
	Barcode        *string
	Tags           []string
	Attributes     []byte
}
