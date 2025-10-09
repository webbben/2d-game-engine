package entity

type CurMax struct {
	CurrentVal int
	MaxVal     int
}

type Vitals struct {
	Health  CurMax
	Stamina CurMax
}

type Attributes struct {
	Strength  CurMax
	Endurance CurMax
	Agility   CurMax
}

func (e Entity) GetMaxInventoryWeight() int {
	return e.Attributes.Strength.CurrentVal * 10
}
