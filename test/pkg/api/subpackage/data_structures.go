package subpackage

type SimpleStructure struct {
	Id    int    `json:"id" swaggo:"true,the user id,2"`
	Name  string `json:"name" swaggo:",the user name,John Smith"`
	Phone string `from:"phone" swaggo:"false,This is to test the from tag"`
}
