package godag

type GoGraph struct {
	NodesMap map[string]GNode
}

type GNode struct {
	Data     interface{}
	Children []interface{}
	Parents  []interface{}
}

func NewGraph() *GoGraph {
	nodesMap := make(map[string]GNode)
	return &GoGraph{NodesMap: nodesMap}
}

func (graph *GoGraph) AddVertex(id string, data interface{}) {
	graph.NodesMap[id] = GNode{Data: data}
}

func (graph *GoGraph) GetVertex(id string) interface{} {
	return graph.NodesMap[id].Data
}

func (graph *GoGraph) AddEdge(from string, to string) {
	fromNode := graph.NodesMap[from]
	toNode := graph.NodesMap[to]

	fromNode.Children = append(fromNode.Children, toNode)
	toNode.Parents = append(toNode.Parents, fromNode)
}

func (graph *GoGraph) GetChildren(id string) []interface{} {
	return graph.NodesMap[id].Children
}

func (graph *GoGraph) GetParents(id string) []interface{} {
	return graph.NodesMap[id].Parents
}

/*
func (graph *GoGraph) removeVertext(id string) {
	gNode := graph.NodesMap[id]
	parents := gNode.Parents
	children := gNode.Children

	for _, parent := range parents {

	}

	for _, child := range children {

	}

	delete(graph.NodesMap, id)
}

func delElementInSlice(gNodes []GNode, delgNode GNode) []GNode {
	delIndex := -1

	//sList := sockList
	for i, gNode := range gNodes {
		if gNode == delgNode {
			delIndex = i
		}
	}

	if delIndex != -1 {
		sockList = append(gNodes[:delIndex], gNodes[delIndex+1:]...)
	}

	return gNodes
}*/
