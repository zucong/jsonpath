package jsonpath

import (
	"fmt"
)

type NewJsonpath struct {
	name       string
	parser     *Parser
	writeMode  bool
	dataHolder []interface{}
	warnings   []string
}

func New(name string, expr string) (*NewJsonpath, error) {
	j := &NewJsonpath{
		name: name,
	}
	p, err := Parse(j.name, "{"+expr+"}")
	if err != nil {
		return nil, fmt.Errorf("cannot parse jsonpath string")
	}
	j.parser = p
	return j, nil
}

func (j *NewJsonpath) AddWarning(warning string) {
	j.warnings = append(j.warnings, warning)
}

func (j *NewJsonpath) InitData(obj interface{}) {
	j.dataHolder = append(j.dataHolder, obj)
}

func (j *NewJsonpath) Data() interface{} {
	return j.dataHolder[0]
}

func (j *NewJsonpath) FindResult() ([]Footprint, error) {
	if j.parser == nil {
		return nil, fmt.Errorf("%s is an incomplete jsonpath expr", j.name)
	}

	var i interface{}
	i = j.dataHolder
	fp := NewFootprint(&i, nil)
	selected, err := fp.SelectAll()
	if err != nil {
		return nil, err
	}

	node := j.parser.Root.Nodes[0]
	footprints, err := j.evalList([]Footprint{selected}, node.(*ListNode))
	if err != nil {
		return nil, err
	}
	return footprints, nil
}

func (j *NewJsonpath) Get() ([]interface{}, error) {
	j.writeMode = false
	footprints, err := j.FindResult()
	if err != nil {
		return []interface{}{}, err
	}
	result := make([]interface{}, 0)
	footprints = expandFootprints(footprints, true)
	for _, footprint := range footprints {
		result = append(result, footprint.HolderPtr())
	}
	return result, nil
}

func (j *NewJsonpath) Set(data *interface{}, change interface{}) error {
	j.writeMode = true
	footprints, err := j.FindResult()
	if err != nil {
		return err
	}

	for _, footprint := range footprints {
		err := footprint.UpdateAll(change)
		if err != nil {
			return err
		}
	}
	return nil
}

func (j *NewJsonpath) walk(footprints []Footprint, node Node) ([]Footprint, error) {
	switch node := node.(type) {
	case *ListNode:
		return j.evalList(footprints, node)
	case *FieldNode:
		return j.evalField(footprints, node)
	case *ArrayNode:
		return j.evalArray(footprints, node)
	case *IntNode:
		return j.evalInt(footprints, node)
	case *BoolNode:
		return j.evalBool(footprints, node)
	case *FloatNode:
		return j.evalFloat(footprints, node)
	case *WildcardNode:
		return j.evalWildcard(footprints, node)
	case *RecursiveNode:
		return j.evalRecursive(footprints, node)
	case *UnionNode:
		return j.evalUnion(footprints, node)
	case *FilterNode:
		return j.evalFilter(footprints, node)
	case *ArrayElementNode:
		return j.evalArrayElement(footprints, node)
	default:
		return footprints, fmt.Errorf("unexpected Node %v", node)
	}
}
