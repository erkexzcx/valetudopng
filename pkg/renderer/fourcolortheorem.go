package renderer

import (
	"image/color"
	"sort"
)

type vertex struct {
	id       string
	adjacent map[string]struct{}
	color    int
}

type graph struct {
	vertices map[string]*vertex
}

func newVertex(id string) *vertex {
	return &vertex{id: id, adjacent: make(map[string]struct{}), color: -1}
}

func (g *graph) addEdge(id1, id2 string) {
	if g.vertices == nil {
		g.vertices = make(map[string]*vertex)
	}

	if _, ok := g.vertices[id1]; !ok {
		g.vertices[id1] = newVertex(id1)
	}

	if _, ok := g.vertices[id2]; !ok {
		g.vertices[id2] = newVertex(id2)
	}

	g.vertices[id1].adjacent[id2] = struct{}{}
	g.vertices[id2].adjacent[id1] = struct{}{}
}

func nextAvailableColor(colors map[int]struct{}) int {
	for color := 0; ; color++ {
		if _, exists := colors[color]; !exists {
			return color
		}
	}
}

func (g *graph) colorVertices() {
	vertices := make([]*vertex, 0, len(g.vertices))
	for _, v := range g.vertices {
		vertices = append(vertices, v)
	}

	sort.Slice(vertices, func(i, j int) bool {
		if len(vertices[i].adjacent) == len(vertices[j].adjacent) {
			return vertices[i].id < vertices[j].id
		}
		return len(vertices[i].adjacent) > len(vertices[j].adjacent)
	})

	for _, v := range vertices {
		if len(v.adjacent) == 0 {
			v.color = 0
		} else {
			colors := make(map[int]struct{})
			for id := range v.adjacent {
				vertex := g.vertices[id]
				if vertex.color != -1 {
					colors[vertex.color] = struct{}{}
				}
			}
			v.color = nextAvailableColor(colors)
		}
	}
}

var fourColors = []color.RGBA{
	{R: 25, G: 161, B: 161, A: 255}, // Teal
	{R: 122, G: 192, B: 55, A: 255}, // Light Green
	{R: 255, G: 155, B: 87, A: 255}, // Orange
	{R: 247, G: 200, B: 65, A: 255}, // Light Yellow
}

func (vi *valetudoImage) findFourColors() {
	g := graph{}

	for _, layer := range vi.layers["segment"] {
		for _, otherLayer := range vi.layers["segment"] {
			if layer != otherLayer && areAdjacent(layer, otherLayer, vi.valetudoJSON.PixelSize) {
				g.addEdge(layer.MetaData.SegmentId, otherLayer.MetaData.SegmentId)
			}
		}
	}

	g.colorVertices()

	for _, v := range g.vertices {
		vi.segmentColor[v.id] = fourColors[v.color]
	}
}

func areAdjacent(layer1, layer2 *Layer, threshold int) bool {
	return (layer1.Dimensions.X.Max >= layer2.Dimensions.X.Min-threshold) &&
		(layer1.Dimensions.X.Min <= layer2.Dimensions.X.Max+threshold) &&
		(layer1.Dimensions.Y.Max >= layer2.Dimensions.Y.Min-threshold) &&
		(layer1.Dimensions.Y.Min <= layer2.Dimensions.Y.Max+threshold)
}
