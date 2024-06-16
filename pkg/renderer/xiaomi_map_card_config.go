package renderer

import (
	"bytes"

	"gopkg.in/yaml.v2"
)

// XiaomiMapCardConfig is based on https://github.com/PiotrMachowski/lovelace-xiaomi-vacuum-map-card/blob/master/docs/demo_config.yaml
type XiaomiMapCardConfig struct {
	Type              string             `yaml:"type"`
	Title             string             `yaml:"title"`
	PresetName        string             `yaml:"preset_name"`
	Entity            string             `yaml:"entity"`
	MapLocked         bool               `yaml:"map_locked"`
	TwoFingerPan      bool               `yaml:"two_finger_pan"`
	MapSource         MapSource          `yaml:"map_source"`
	InternalVariables InternalVariables  `yaml:"internal_variables"`
	CalibrationSource CalibrationSource  `yaml:"calibration_source"`
	MapModes          []MapMode          `yaml:"map_modes"`
	AdditionalPresets []AdditionalPreset `yaml:"additional_presets,omitempty"`
	segments          []segment
}

type MapSource struct {
	Camera string `yaml:"camera"`
}

type CalibrationSource struct {
	Camera *bool  `yaml:"camera,omitempty"`
	Entity string `yaml:"entity,omitempty"`
}

type InternalVariables struct {
	Topics string `yaml:"topics,omitempty"`
}

type MapMode struct {
	Template             string                `yaml:"template,omitempty"`
	PredefinedSelections []PredefinedSelection `yaml:"predefined_selections,omitempty"`
}

type PredefinedSelection struct {
	Zones    [][]int `yaml:"zones,omitempty"`
	Label    Label   `yaml:"label"`
	Icon     Icon    `yaml:"icon"`
	ID       string  `yaml:"id,omitempty"`
	Outline  [][]int `yaml:"outline,omitempty"`
	Position []int   `yaml:"position,omitempty"`
}

type Label struct {
	Text    string `yaml:"text"`
	X       int    `yaml:"x"`
	Y       int    `yaml:"y"`
	OffsetY int    `yaml:"offset_y"`
}

type Icon struct {
	Name string `yaml:"name"`
	X    int    `yaml:"x,omitempty"`
	Y    int    `yaml:"y,omitempty"`
}
type AdditionalPreset struct {
	Name                 string                `yaml:"name"`
	Icon                 string                `yaml:"icon"`
	SelectionType        string                `yaml:"selection_type"`
	MaxSelections        int                   `yaml:"max_selections"`
	RepeatsType          string                `yaml:"repeats_type"`
	MaxRepeats           int                   `yaml:"max_repeats"`
	ServiceCallSchema    ServiceCallSchema     `yaml:"service_call_schema"`
	PredefinedSelections []PredefinedSelection `yaml:"predefined_selections,omitempty"`
}

type ServiceCallSchema struct {
	Service     string      `yaml:"service"`
	ServiceData ServiceData `yaml:"service_data"`
	Target      Target      `yaml:"target"`
}

type ServiceData struct {
	Path       string `yaml:"path,omitempty"`
	Repeats    string `yaml:"repeats,omitempty"`
	Predefined string `yaml:"predefined,omitempty"`
	Point      string `yaml:"point,omitempty"`
	PointX     string `yaml:"point_x,omitempty"`
	PointY     string `yaml:"point_y,omitempty"`
}

type Target struct {
	EntityID string `yaml:"entity_id"`
}

type segment struct {
	name string
	id   string
	d    Dimensions
}

func newMapConf(r *Renderer) *XiaomiMapCardConfig {
	name := r.conf.Mqtt.Topics.ValetudoIdentifier
	return &XiaomiMapCardConfig{
		Type:              "custom:xiaomi-vacuum-map-card",
		Title:             "Xiaomi Vacuum Map Card",
		PresetName:        "Live map",
		Entity:            "vacuum.valetudo_" + name,
		MapSource:         MapSource{"camera." + name + "_map"},
		CalibrationSource: CalibrationSource{Entity: "sensor." + name + "_calibration"},
		InternalVariables: InternalVariables{"valetudo/" + name},
		segments:          make([]segment, 0, 3),
		MapLocked:         true,
	}
}

func (x *XiaomiMapCardConfig) addSegment(m *MetaData, d Dimensions) {
	if m == nil || m.SegmentId == "" || m.Name == "" {
		return
	}
	x.segments = append(x.segments, segment{
		name: m.Name,
		id:   m.SegmentId,
		d:    d,
	})
}

func (x *XiaomiMapCardConfig) setMapModes() {
	x.MapModes = make([]MapMode, 0, 4)

	if len(x.segments) > 0 {
		m := MapMode{
			Template:             "vacuum_clean_segment",
			PredefinedSelections: make([]PredefinedSelection, 0, len(x.segments)),
		}

		for _, s := range x.segments {
			m.PredefinedSelections = append(m.PredefinedSelections, PredefinedSelection{
				ID: s.id,
				Label: Label{
					Text: s.name,
					X:    s.d.X.Mid*5 - 20,
					Y:    s.d.Y.Mid*5 - 50,
				},
				Icon: Icon{
					Name: "mdi:broom",
					X:    s.d.X.Mid * 5,
					Y:    s.d.Y.Mid * 5,
				},
			})
		}

		x.MapModes = append(x.MapModes, m)
	}

	x.MapModes = append(x.MapModes,
		MapMode{Template: "vacuum_goto"},
		MapMode{Template: "vacuum_clean_zone"},
		MapMode{Template: "vacuum_goto_predefined"},
	)
}

func (x *XiaomiMapCardConfig) asYaml() []byte {
	reqBodyBytes := new(bytes.Buffer)
	yaml.NewEncoder(reqBodyBytes).Encode(x)
	return reqBodyBytes.Bytes()
}
