package properties

import (
	"encoding/json"

	"github.com/webbben/2d-game-engine/logz"
)

// Property represents a custom property from a Tiled map/tileset JSON file.
type Property struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"` // string, int, float, bool, color, file, object, class
	Value        any    `json:"value"`
	PropertyType string `json:"propertytype,omitempty"` // For object/class types
}

// UnmarshalJSON handles the flexible property value types
func (p *Property) UnmarshalJSON(data []byte) error {
	type Alias Property
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Set default type if not specified
	if p.Type == "" {
		p.Type = "string"
	}

	return nil
}

// GetStringValue returns the property value as a string
func (p *Property) GetStringValue() string {
	if str, ok := p.Value.(string); ok {
		return str
	}
	return ""
}

// GetIntValue returns the property value as an int
func (p *Property) GetIntValue() int {
	switch v := p.Value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	}
	return 0
}

// GetFloatValue returns the property value as a float64
func (p *Property) GetFloatValue() float64 {
	switch v := p.Value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}
	return 0.0
}

// GetBoolValue returns the property value as a bool
func (p *Property) GetBoolValue() bool {
	if b, ok := p.Value.(bool); ok {
		return b
	}
	return false
}

// GetBoolProperty returns the value of the named property as a bool.
// The second return value is false if the property wasn't found.
func GetBoolProperty(propName string, props []Property) (value, found bool) {
	for _, prop := range props {
		if prop.Name == propName {
			return prop.GetBoolValue(), true
		}
	}

	return false, false
}

// GetStringProperty returns the value of the named property as a string.
// The second return value is false if the property wasn't found.
func GetStringProperty(propName string, props []Property) (val string, found bool) {
	for _, prop := range props {
		if prop.Name == propName {
			return prop.GetStringValue(), true
		}
	}

	return "", false
}

// GetFloatProperty returns the value of the named property as a float64.
// The second return value is false if the property wasn't found.
func GetFloatProperty(propName string, props []Property) (val float64, found bool) {
	for _, prop := range props {
		if prop.Name == propName {
			return prop.GetFloatValue(), true
		}
	}

	return 0, false
}

// GetIntProperty returns the value of the named property as an int.
// The second return value is false if the property wasn't found.
func GetIntProperty(propName string, props []Property) (val int, found bool) {
	for _, prop := range props {
		if prop.Name == propName {
			return prop.GetIntValue(), true
		}
	}

	return 0, false
}

type LightProps struct {
	R, G, B           float64 // must be between 0 and 1
	GlowFactor        float64
	InnerRadiusFactor float64
	OffsetY           int
	Radius            int
	FlickerInterval   int
	MaxBrightness     float64
	CoreRadiusFactor  float64
}

type WindowProps struct {
	Length       int
	Width        int
	MaxIntensity float64
	DirX, DirY   int
}

// GetLightProps reads the light-related properties from a property list.
func GetLightProps(p []Property) LightProps {
	props := LightProps{}

	for _, prop := range p {
		switch prop.Name {
		case PropLightColorR:
			props.R = prop.GetFloatValue()
		case PropLightColorG:
			props.G = prop.GetFloatValue()
		case PropLightColorB:
			props.B = prop.GetFloatValue()
		case PropLightGlowFactor:
			props.GlowFactor = prop.GetFloatValue()
		case PropLightOffsetY:
			props.OffsetY = prop.GetIntValue()
		case PropLightRadius:
			props.Radius = prop.GetIntValue()
		case PropLightInnerRadiusFactor:
			props.InnerRadiusFactor = prop.GetFloatValue()
		case PropLightFlickerInterval:
			props.FlickerInterval = prop.GetIntValue()
		case PropLightMaxBrightness:
			props.MaxBrightness = prop.GetFloatValue()
		case PropLightCoreRadius:
			props.CoreRadiusFactor = prop.GetFloatValue()
		case PropLightPreset:
			logz.Panicln("GetLightProps", "light_preset prop is deprecated; use light_color_r/g/b props instead.")
		}
	}

	return props
}

// GetWindowProps reads the window-related properties from a property list.
func GetWindowProps(p []Property) WindowProps {
	props := WindowProps{}

	for _, prop := range p {
		switch prop.Name {
		case PropWindowWidth:
			props.Width = prop.GetIntValue()
		case PropWindowLength:
			props.Length = prop.GetIntValue()
		case PropWindowMaxIntensity:
			props.MaxIntensity = prop.GetFloatValue()
		case PropWindowDirX:
			props.DirX = prop.GetIntValue()
		case PropWindowDirY:
			props.DirY = prop.GetIntValue()
		}
	}

	return props
}

