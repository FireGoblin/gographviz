//Copyright 2013 Vastech SA (PTY) LTD
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package gographviz

import "github.com/firegoblin/gographviz/ast"
import "strings"
import "fmt"
import "os"

type GraphableNode interface {
	NodeInterface
	EdgesInterface
}

var unsafeNames = []string{"graph", "node", "edge", "subgraph", "digraph"}

func SafeName(s string) string {
	lower := strings.ToLower(s)
	for _, v := range unsafeNames {
		if lower == v {
			return s + "TYPE"
		}
	}

	return s
}

//The analysed representation of the Graph parsed from the DOT format.
type Graph struct {
	Attrs     Attrs
	Name      string
	Directed  bool
	Strict    bool
	Nodes     *Nodes
	Edges     *Edges
	SubGraphs *SubGraphs
	Relations *Relations
}

//Creates a new empty graph, ready to be populated.
func NewGraph() *Graph {
	return &Graph{
		Attrs:     make(Attrs),
		Name:      "",
		Directed:  false,
		Strict:    false,
		Nodes:     NewNodes(),
		Edges:     NewEdges(),
		SubGraphs: NewSubGraphs(),
		Relations: NewRelations(),
	}
}

//If the graph is strict then multiple edges are not allowed between the same pairs of nodes,
//see dot man page.
func (this *Graph) SetStrict(strict bool) {
	this.Strict = strict
}

//Sets whether the graph is directed (true) or undirected (false).
func (this *Graph) SetDir(dir bool) {
	this.Directed = dir
}

//Sets the graph name.
func (this *Graph) SetName(name string) {
	this.Name = name
}

//Adds an edge to the graph from node src to node dst.
//srcPort and dstPort are the port the node ports, leave as empty strings if it is not required.
//This does not imply the adding of missing nodes.
func (this *Graph) AddPortEdge(src, srcPort, dst, dstPort string, directed bool, attrs map[string]string) {
	this.Edges.Add(&Edge{src, srcPort, dst, dstPort, directed, attrs})
}

//Adds an edge to the graph from node src to node dst.
//This does not imply the adding of missing nodes.
func (this *Graph) AddEdge(src, dst string, directed bool, attrs map[string]string) {
	this.AddPortEdge(src, "", dst, "", directed, attrs)
}

func (this *Graph) AddEdgesInterface(edges EdgesInterface) {
	for _, v := range edges.Edges() {
		this.Edges.Add(v)
	}
}

func (this *Graph) RemoveEdgelessNodes(parentGraph string) {
	for key := range this.Nodes.Lookup {
		_, ok := this.Edges.SrcToDsts[key]
		if !ok {
			found := false
			for _, val := range this.Edges.SrcToDsts {
				_, ok := val[key]
				if ok {
					found = true
					break
				}
			}

			if !found {
				fmt.Fprintln(os.Stderr, "removing the node:", key)
				this.RemoveNode(parentGraph, key)
			}
		}
	}
}

//Adds a node to a graph/subgraph.
//If not subgraph exists use the name of the main graph.
//This does not imply the adding of a missing subgraph.
func (this *Graph) AddNode(parentGraph string, name string, attrs map[string]string) {
	this.Nodes.Add(&Node{name, attrs})
	this.Relations.Add(parentGraph, name)
}

func (this *Graph) RemoveNode(parentGraph string, name string) {
	this.Nodes.Remove(name)
	this.Relations.Remove(parentGraph, name)
}

func (this *Graph) AddNodeInterface(parentGraph string, node NodeInterface) {
	this.Nodes.Add(&Node{node.Name(), node.Attrs()})
	this.Relations.Add(parentGraph, node.Name())
}

func (this *Graph) AddGraphableNode(parentGraph string, node GraphableNode) {
	this.AddNodeInterface(parentGraph, node)
	this.AddEdgesInterface(node)
}

func (this *Graph) AddGraphableNodes(parentGraph string, nodes []GraphableNode) {
	for _, n := range nodes {
		this.AddGraphableNode(parentGraph, n)
	}
}

func (this *Graph) getAttrs(graphName string) Attrs {
	if this.Name == graphName {
		return this.Attrs
	}
	g, ok := this.SubGraphs.SubGraphs[graphName]
	if !ok {
		panic("graph or subgraph " + graphName + " does not exist")
	}
	return g.Attrs
}

//Adds an attribute to a graph/subgraph.
func (this *Graph) AddAttr(parentGraph string, field string, value string) {
	this.getAttrs(parentGraph).Add(field, value)
}

//Adds a subgraph to a graph/subgraph.
func (this *Graph) AddSubGraph(parentGraph string, name string, attrs map[string]string) {
	this.SubGraphs.Add(name)
	for key, value := range attrs {
		this.AddAttr(name, key, value)
	}
}

func (this *Graph) IsNode(name string) bool {
	_, ok := this.Nodes.Lookup[name]
	return ok
}

func (this *Graph) IsSubGraph(name string) bool {
	_, ok := this.SubGraphs.SubGraphs[name]
	return ok
}

//Creates an Abstract Syntrax Tree from the Graph.
func (g *Graph) WriteAst() *ast.Graph {
	w := newWriter(g)
	return w.Write()
}

//Returns a DOT string representing the Graph.
func (g *Graph) String() string {
	return g.WriteAst().String()
}
