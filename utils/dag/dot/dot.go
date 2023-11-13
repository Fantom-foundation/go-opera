/*
Package dot implements an API to produce Graphviz dot language output.

Basic Graph creation:

	g := dot.NewGraph("G")
	g.SetType(dot.DIGRAPH)
	...
	g.AddEdge(dot.NewNode("A"), dot.NewNode("B"))
	...
	fmt.Sprint(g)
*/
package dot

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var AttributeError = errors.New("Invalid Attribute")

var graphAttributes = []string{"Damping", "K", "URL", "aspect", "bb", "bgcolor",
	"center", "charset", "clusterrank", "colorscheme", "comment", "compound",
	"concentrate", "defaultdist", "dim", "dimen", "diredgeconstraints",
	"dpi", "epsilon", "esep", "fontcolor", "fontname", "fontnames",
	"fontpath", "fontsize", "id", "label", "labeljust", "labelloc",
	"landscape", "layers", "layersep", "layout", "levels", "levelsgap",
	"lheight", "lp", "lwidth", "margin", "maxiter", "mclimit", "mindist",
	"mode", "model", "mosek", "nodesep", "nojustify", "normalize", "nslimit",
	"nslimit1", "ordering", "orientation", "outputorder", "overlap",
	"overlap_scaling", "pack", "packmode", "pad", "page", "pagedir",
	"quadtree", "quantum", "rankdir", "ranksep", "ratio", "remincross",
	"repulsiveforce", "resolution", "root", "rotate", "searchsize", "sep",
	"showboxes", "style", "size", "smoothing", "sortv", "splines", "start",
	"stylesheet", "target", "truecolor", "viewport", "voro_margin",
	"rank", "newrank"}

var edgeAttributes = []string{"URL", "arrowhead", "arrowsize", "arrowtail",
	"color", "colorscheme", "comment", "constraint", "decorate", "dir",
	"edgeURL", "edgehref", "edgetarget", "edgetooltip", "fontcolor",
	"fontname", "fontsize", "headURL", "headclip", "headhref", "headlabel",
	"headport", "headtarget", "headtooltip", "href", "id", "label",
	"labelURL", "labelangle", "labeldistance", "labelfloat", "labelfontcolor",
	"labelfontname", "labelfontsize", "labelhref", "labeltarget",
	"labeltooltip", "layer", "len", "lhead", "lp", "ltail", "minlen",
	"nojustify", "penwidth", "pos", "samehead", "sametail", "showboxes",
	"style", "tailURL", "tailclip", "tailhref", "taillabel", "tailport",
	"tailtarget", "tailtooltip", "target", "tooltip", "weight",
	// for subgraphs
	"rank"}

var nodeAttributes = []string{"URL", "color", "colorscheme", "comment",
	"distortion", "fillcolor", "fixedsize", "fontcolor", "fontname",
	"fontsize", "group", "height", "id", "image", "imagescale", "label",
	"labelloc", "layer", "margin", "nojustify", "orientation", "penwidth",
	"peripheries", "pin", "pos", "rects", "regular", "root", "samplepoints",
	"shape", "shapefile", "showboxes", "sides", "skew", "sortv", "style",
	"target", "tooltip", "vertices", "width", "z",
	// The following are attributes dot2tex
	"texlbl", "texmode"}

var clusterAttributes = []string{"K", "URL", "bgcolor", "color", "colorscheme",
	"fillcolor", "fontcolor", "fontname", "fontsize", "label", "labeljust",
	"labelloc", "lheight", "lp", "lwidth", "nojustify", "pencolor",
	"penwidth", "peripheries", "sortv", "style", "target", "tooltip"}

var dotKeywords = []string{"graph", "subgraph", "digraph", "node", "edge", "strict"}

type GraphType int

const (
	DIGRAPH GraphType = iota
	GRAPH
	SUBGRAPH
)

// Fields common to all graph object types
type common struct {
	_type       string
	name        string
	attributes  map[string]string
	sequence    int
	parentGraph *Graph
}

type GraphObject interface {
	Type() string
	Get(string) string
	Set(string, string) error
	GetParentGraph() *Graph
	SetParentGraph(g *Graph)
	Sequence() int
}

type graphObjects []GraphObject

func (gol graphObjects) Len() int {
	return len(gol)
}

func (gol graphObjects) Less(i, j int) bool {
	return gol[i].Sequence() < gol[j].Sequence()
}

func (gol graphObjects) Swap(i, j int) {
	gol[i], gol[j] = gol[j], gol[i]
}

type Graph struct {
	common
	nodeAttributes       map[string]string
	edgeAttributes       map[string]string
	sameRank             [][]string
	strict               bool
	graphType            GraphType
	supressDisconnected  bool
	simplify             bool
	currentChildSequence int
	nodes                map[string][]*Node
	edges                map[string][]*Edge
	subgraphs            map[string][]*SubGraph
}

func NewGraph(name string) *Graph {
	g := &Graph{
		common: common{
			_type:      "graph",
			name:       name,
			attributes: make(map[string]string, 0),
		},
		nodeAttributes:       make(map[string]string),
		edgeAttributes:       make(map[string]string),
		sameRank:             make([][]string, 0),
		nodes:                make(map[string][]*Node, 0),
		edges:                make(map[string][]*Edge, 0),
		subgraphs:            make(map[string][]*SubGraph, 0),
		currentChildSequence: 1,
	}
	g.SetParentGraph(g)
	return g
}

type SubGraph struct {
	Graph
}

func NewSubgraph(name string) *SubGraph {
	result := &SubGraph{
		Graph: *NewGraph(name),
	}
	result._type = "subgraph"
	result.graphType = SUBGRAPH
	return result
}

func indexInSlice(slice []string, toFind string) int {
	for i, v := range slice {
		if v == toFind {
			return i
		}
	}
	return -1
}

var alreadyQuotedRegex = regexp.MustCompile("^\".+\"$")
var validIdentifierRegexWithPort = regexp.MustCompile("^[_a-zA-Z][a-zA-Z0-9_,:\"]*[a-zA-Z0-9_,\"]+$")
var validIdentifierRegex = regexp.MustCompile("^[_a-zA-Z][a-zA-Z0-9_,]*$")

func needsQuotes(s string) bool {
	if indexInSlice(dotKeywords, s) != -1 {
		return false
	}
	if alreadyQuotedRegex.MatchString(s) {
		return false
	}
	if validIdentifierRegexWithPort.MatchString(s) || validIdentifierRegex.MatchString(s) {
		return false
	}

	return true
}

func QuoteIfNecessary(s string) (result string) {
	if needsQuotes(s) {
		s = strings.Replace(s, "\"", "\\\"", -1)
		s = strings.Replace(s, "\n", "\\n", -1)
		s = strings.Replace(s, "\r", "\\r", -1)
		s = "\"" + s + "\""
	}
	return s
}

func validAttribute(attributeCollection []string, attributeName string) bool {
	return indexInSlice(attributeCollection, attributeName) != -1
}

func validGraphAttribute(attributeName string) bool {
	return validAttribute(graphAttributes, attributeName)
}

func validNodeAttribute(attributeName string) bool {
	return validAttribute(nodeAttributes, attributeName)
}

func sortedKeys(sourceMap map[string]string) []string {
	keys := make([]string, 0, len(sourceMap))
	for k, _ := range sourceMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

////////////////////////////////////////////////////////////////////////////////////////

func (gt GraphType) String() string {
	if gt == DIGRAPH {
		return "digraph"
	} else if gt == GRAPH {
		return "graph"
	} else if gt == SUBGRAPH {
		return "subgraph"
	}
	return "(invalid)"
}

func (c *common) Type() string {
	return c._type
}

func (c *common) GetParentGraph() *Graph {
	return c.parentGraph
}

func (c *common) SetParentGraph(g *Graph) {
	c.parentGraph = g
}

func (c *common) Sequence() int {
	return c.sequence
}

func (c *common) Get(attributeName string) string {
	return c.attributes[attributeName]
}

func (c *common) Set(attributeName, attributeValue string) error {
	c.attributes[attributeName] = attributeValue
	return nil
}

func setAttribute(validAttributes []string, attributes map[string]string, attributeName, attributeValue string) error {
	if validAttribute(validAttributes, attributeName) {
		attributes[attributeName] = attributeValue
		return nil
	}
	return AttributeError
}

func (g *Graph) Set(attributeName, attributeValue string) error {
	return setAttribute(graphAttributes, g.common.attributes, attributeName, attributeValue)
}

func (g *Graph) SetGlobalNodeAttr(attributeName, attributeValue string) error {
	return setAttribute(nodeAttributes, g.nodeAttributes, attributeName, attributeValue)
}

func (g *Graph) SetGlobalEdgeAttr(attributeName, attributeValue string) error {
	return setAttribute(edgeAttributes, g.edgeAttributes, attributeName, attributeValue)
}

func (n *Node) Set(attributeName, attributeValue string) error {
	return setAttribute(nodeAttributes, n.common.attributes, attributeName, attributeValue)
}

func (e *Edge) Set(attributeName, attributeValue string) error {
	return setAttribute(edgeAttributes, e.common.attributes, attributeName, attributeValue)
}

func (c *common) setSequence(sequence int) {
	c.sequence = sequence
}

// SameRank enforces alignment of the given nodes
func (g *Graph) SameRank(nodes []string) {
	g.sameRank = append(g.sameRank, nodes)
}

// Set the type of the graph, valid values are GRAPH or DIGRAPH
func (g *Graph) SetType(t GraphType) {
	g.graphType = t
	// @todo consider disallowing setting type to SUBGRAPH
}

func (c common) Name() string {
	return c.name
}

func (g *Graph) GetRoot() (result *Graph) {
	result = g
	for parent := g.GetParentGraph(); parent != result; parent = parent.GetParentGraph() {
		result = parent
	}
	return result
}

func (g *Graph) getNextSequenceNumber() (next int) {
	next = g.currentChildSequence
	g.currentChildSequence += 1
	return
}
func (g *Graph) AddNode(n *Node) {
	name := n.Name()
	if _, ok := g.nodes[name]; !ok {
		g.nodes[name] = make([]*Node, 0)
	}
	n.setSequence(g.getNextSequenceNumber())
	n.SetParentGraph(g.GetParentGraph())
	g.nodes[name] = append(g.nodes[name], n)
}

func (g *Graph) AddEdge(e *Edge) {
	name := e.Name()
	if _, ok := g.edges[name]; !ok {
		g.edges[name] = make([]*Edge, 0)
	}
	e.setSequence(g.getNextSequenceNumber())
	e.SetParentGraph(g.GetParentGraph())
	g.edges[name] = append(g.edges[name], e)
}

func (g *Graph) AddSubgraph(sg *SubGraph) {
	name := sg.Name()
	if _, ok := g.subgraphs[name]; !ok {
		g.subgraphs[name] = make([]*SubGraph, 0)
	}
	sg.setSequence(g.getNextSequenceNumber())
	g.subgraphs[name] = append(g.subgraphs[name], sg)
}

func (g *Graph) GetSubgraphs() (result []*SubGraph) {
	result = make([]*SubGraph, 0)
	for _, sgs := range g.subgraphs {
		for _, sg := range sgs {
			result = append(result, sg)
		}
	}
	return result
}

func (g Graph) String() string {
	var parts []string
	if g.strict {
		parts = append(parts, "strict ")
	}
	if g.name == "" {
		parts = append(parts, "{\n")
	} else {
		parts = append(parts, fmt.Sprintf("%s %s {\n", g.graphType, QuoteIfNecessary(g.name)))
	}

	if len(g.attributes) > 0 {
		attrs := make([]string, 0, len(g.attributes))
		for _, key := range sortedKeys(g.attributes) {
			attrs = append(attrs, "  "+key+"="+QuoteIfNecessary(g.attributes[key]))
		}
		if len(attrs) > 0 {
			parts = append(parts, "graph [\n")
			parts = append(parts, strings.Join(attrs, ";\n"))
			parts = append(parts, ";\n];\n")
		}
	}

	if len(g.nodeAttributes) > 0 {
		attrs := make([]string, 0, len(g.nodeAttributes))
		for _, key := range sortedKeys(g.nodeAttributes) {
			attrs = append(attrs, "  "+key+"="+QuoteIfNecessary(g.nodeAttributes[key]))
		}
		if len(attrs) > 0 {
			parts = append(parts, "node [\n")
			parts = append(parts, strings.Join(attrs, ";\n"))
			parts = append(parts, ";\n];\n")
		}
	}

	if len(g.edgeAttributes) > 0 {
		attrs := make([]string, 0, len(g.edgeAttributes))
		for _, key := range sortedKeys(g.edgeAttributes) {
			attrs = append(attrs, "  "+key+"="+QuoteIfNecessary(g.edgeAttributes[key]))
		}
		if len(attrs) > 0 {
			parts = append(parts, "edge [\n")
			parts = append(parts, strings.Join(attrs, ";\n"))
			parts = append(parts, ";\n];\n")
		}
	}

	objectList := make(graphObjects, 0)

	for _, nodes := range g.nodes {
		for _, node := range nodes {
			objectList = append(objectList, node)
		}
	}
	for _, edges := range g.edges {
		for _, edge := range edges {
			objectList = append(objectList, edge)
		}
	}
	for _, subgraphs := range g.subgraphs {
		for _, subgraph := range subgraphs {
			objectList = append(objectList, subgraph)
		}
	}
	sort.Sort(objectList)

	for _, obj := range objectList {
		//@todo type-based decision making re: supressDisconnected and simplify
		//switch o := obj.(type) {
		//case *Node:
		//}
		parts = append(parts, fmt.Sprintf("%s\n", obj))
	}

	for _, nodes := range g.sameRank {
		parts = append(parts, fmt.Sprintf("{ rank=same %s }", strings.Join(nodes, " ")))
	}

	parts = append(parts, "}\n")
	return strings.Join(parts, "")
}

type Node struct {
	common
}

func NewNode(name string) *Node {
	return &Node{
		common{
			name:       name,
			attributes: make(map[string]string, 0),
		},
	}
}

func (n Node) String() string {

	name := QuoteIfNecessary(n.name)

	parts := make([]string, 0)

	attrs := make([]string, 0)
	for _, key := range sortedKeys(n.attributes) {
		value := n.attributes[key]
		if key == "label" && len(value) > 4 && value[0] == '<' && value[len(value)-1] == '>' {
			attrs = append(attrs, key+"="+value)
		} else {
			attrs = append(attrs, key+"="+QuoteIfNecessary(value))
		}
	}
	if len(attrs) > 0 {
		parts = append(parts, strings.Join(attrs, ", "))
	}

	//@todo don't print if node is empty
	if len(parts) > 0 {
		name += " [" + strings.Join(parts, ", ") + "]"
	}

	return name + ";"
}

type Edge struct {
	common
	points [2]*Node
}

func NewEdge(src, dst *Node) *Edge {
	return &Edge{
		common{
			_type:      "edge",
			attributes: make(map[string]string, 0),
		},
		[2]*Node{src, dst},
	}
}

func (e Edge) Source() *Node {
	return e.points[0]
}

func (e Edge) Destination() *Node {
	return e.points[1]
}

func (e Edge) String() string {
	src, dst := e.Source(), e.Destination()
	parts := make([]string, 0)

	parts = append(parts, QuoteIfNecessary(src.Name()))

	parent := e.GetParentGraph()
	if parent != nil && parent.GetRoot() != nil && parent.GetRoot().graphType == DIGRAPH {
		parts = append(parts, "->")
	} else {
		parts = append(parts, "--")
	}
	parts = append(parts, QuoteIfNecessary(dst.Name()))

	attrs := make([]string, 0)
	for _, key := range sortedKeys(e.attributes) {
		attrs = append(attrs, key+"="+QuoteIfNecessary(e.attributes[key]))
	}
	if len(attrs) > 0 {
		parts = append(parts, " [")
		parts = append(parts, strings.Join(attrs, ", "))
		parts = append(parts, "]")
	}

	return strings.Join(parts, " ")
}

func init() {
	sort.Strings(graphAttributes)
	sort.Strings(nodeAttributes)
	sort.Strings(edgeAttributes)
	sort.Strings(clusterAttributes)
}
