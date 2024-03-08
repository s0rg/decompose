package builder

import (
	"io"

	"github.com/s0rg/decompose/internal/node"
)

type PlantUML struct{}

func NewPlantUML() *PlantUML {
	return &PlantUML{}
}

func (p *PlantUML) Name() string {
	return "plant-uml"
}

func (p *PlantUML) AddNode(n *node.Node) error {
	return nil
}

func (p *PlantUML) AddEdge(e *node.Edge) {

}

func (p *PlantUML) Write(w io.Writer) error {
	return nil
}
