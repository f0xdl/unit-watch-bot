package domain

type Country struct {
	Id   int
	Name string
	Code string
}

type Region struct {
	Id        int
	Name      string
	Code      string
	CountryId int //relation Country
}

type Point struct {
	Id       int
	Name     string
	RegionId int //relation Region
}
