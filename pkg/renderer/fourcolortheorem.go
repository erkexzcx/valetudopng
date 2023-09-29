package renderer

import (
	"image/color"
	"sort"
)

type Vertex struct {
	ID       string
	Adjacent map[string]struct{}
	Color    int
}

type Graph struct {
	Vertices map[string]*Vertex
}

func (g *Graph) AddEdge(id1, id2 string) {
	if g.Vertices == nil {
		g.Vertices = make(map[string]*Vertex)
	}

	v1, ok := g.Vertices[id1]
	if !ok {
		v1 = &Vertex{ID: id1, Adjacent: make(map[string]struct{}), Color: -1}
		g.Vertices[id1] = v1
	}

	v2, ok := g.Vertices[id2]
	if !ok {
		v2 = &Vertex{ID: id2, Adjacent: make(map[string]struct{}), Color: -1}
		g.Vertices[id2] = v2
	}

	v1.Adjacent[id2] = struct{}{}
	v2.Adjacent[id1] = struct{}{}
}

func (g *Graph) ColorVertices() {
	vertices := make([]*Vertex, 0, len(g.Vertices))
	for _, v := range g.Vertices {
		vertices = append(vertices, v)
	}

	sort.Slice(vertices, func(i, j int) bool {
		return len(vertices[i].Adjacent) > len(vertices[j].Adjacent)
	})

	for _, v := range vertices {
		if len(v.Adjacent) == 0 {
			v.Color = 0
		} else {
			colors := make(map[int]struct{})
			for id := range v.Adjacent {
				vertex := g.Vertices[id]
				if vertex.Color != -1 {
					colors[vertex.Color] = struct{}{}
				}
			}

			for color := 0; ; color++ {
				if _, exists := colors[color]; !exists {
					v.Color = color
					break
				}
			}
		}
	}
}

var fourColors = []color.RGBA{
	{R: 25, G: 161, B: 161, A: 255}, // Teal
	{R: 223, G: 86, B: 24, A: 255},  // Dark Orange
	{R: 122, G: 192, B: 55, A: 255}, // Light Green
	{R: 247, G: 200, B: 65, A: 255}, // Light Yellow
}

func (vi *valetudoImage) FindFourColors() {
	graph := Graph{}

	for _, layer := range vi.layers["segment"] {
		for _, otherLayer := range vi.layers["segment"] {
			if layer != otherLayer && areAdjacent(layer, otherLayer, vi.valetudoJSON.PixelSize) {
				graph.AddEdge(layer.MetaData.SegmentId, otherLayer.MetaData.SegmentId)
			}
		}
	}

	graph.ColorVertices()

	for _, v := range graph.Vertices {
		vi.segmentColor[v.ID] = fourColors[v.Color]
	}
}

func areAdjacent(layer1, layer2 *Layer, threshold int) bool {
	return (layer1.Dimensions.X.Max >= layer2.Dimensions.X.Min-threshold) &&
		(layer1.Dimensions.X.Min <= layer2.Dimensions.X.Max+threshold) &&
		(layer1.Dimensions.Y.Max >= layer2.Dimensions.Y.Min-threshold) &&
		(layer1.Dimensions.Y.Min <= layer2.Dimensions.Y.Max+threshold)
}
