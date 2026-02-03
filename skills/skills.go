// Package skills defines skills and attributes structs and concepts
package skills

type CurMax struct {
	CurrentVal int
	MaxVal     int
}

type Vitals struct {
	Health  CurMax
	Stamina CurMax
}

type (
	AttributeID string
	SkillID     string
)

type AttributeDef struct {
	ID          AttributeID
	DisplayName string
}

type SkillDef struct {
	ID          SkillID
	DisplayName string
}
