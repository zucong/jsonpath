package jsonpath

import (
	"fmt"
	"jsonpath/template"
	"log"
)

func expandFootprints(footprints []Footprint, remainUnexpandableFootprint bool) []Footprint {
	if len(footprints) == 0 {
		return footprints
	}
	result := make([]Footprint, 0)
	for _, fp := range footprints {
		fps, err := fp.Expand()
		if err != nil && remainUnexpandableFootprint {
			result = append(result, fp)
		} else {
			result = append(result, fps...)
		}
	}
	return result
}

func (j *Jsonpath) evalList(footprints []Footprint, node *ListNode) ([]Footprint, error) {
	var err error

	for _, n := range node.Nodes {
		footprints, err = j.walk(footprints, n)
		if err != nil {
			return nil, err
		}
	}
	return footprints, nil
}

func (j *Jsonpath) evalField(footprints []Footprint, node *FieldNode) ([]Footprint, error) {
	if j.writeMode {
		for _, footprint := range footprints {
			err := footprint.EnforceObjectSelection()
			if err != nil {
				return nil, err
			}
		}
	}
	footprints = expandFootprints(footprints, false)
	result := make([]Footprint, 0)
	for _, fp := range footprints {
		ref := fp.HolderPtr()
		if m, ok := (*ref).(map[string]interface{}); ok {
			if _, ok := m[node.Value]; ok {
				result = append(result, MapFootprint{
					Ref: ref,
					SelectionKeys: []SelectionKey{{node.Value, VirtualInfo{
						Virtual:  false,
						RealSize: -1,
					}}},
				})
			} else if j.writeMode {
				(*ref).(map[string]interface{})[node.Value] = make(map[string]interface{})
				result = append(result, MapFootprint{
					Ref: ref,
					SelectionKeys: []SelectionKey{{node.Value, VirtualInfo{
						Virtual:  true,
						RealSize: -1,
					}}},
				})
			} else {
				j.AddWarning(fmt.Sprintf("cannot find the field: %s", node.Value))
			}
		}
		//} else {
		//	return nil, fmt.Errorf("cannot use a key string to find a element in a non-map object")
		//}
	}
	return result, nil
}

func (j *Jsonpath) inferArrayNode(arrPtr *[]interface{}, node *ArrayNode) (base, limit, step int, needInvert bool) {
	arr := *arrPtr
	if len(node.Params) == 1 {
		return node.Params[0].Value, node.Params[0].Value + 1, 1, false
	}

	x, y, z := node.Params[0], node.Params[1], node.Params[2]

	// infer step
	if z.Known {
		step = z.Value
	}
	if step == 0 {
		step = 1
	} else if step < 0 {
		needInvert = true
	}

	if x.Value > len(arr)-1 {
		if step < 0 {
			base = len(arr) - 1
		} else {
			base = x.Value
		}
	} else if x.Value >= 0 {
		base = x.Value
	} else if x.Value >= -len(arr) {
		base = x.Value + len(arr)
	} else {
		base = 0
	}

	if y.Value >= 0 {
		limit = y.Value
	} else if y.Value >= -len(arr) {
		limit = y.Value + len(arr)
	} else {
		limit = -1
	}

	if !x.Known {
		if step > 0 {
			base = 0
		} else {
			base = len(arr) - 1
		}
	}

	if !y.Known {
		if step > 0 {
			limit = len(arr)
		} else {
			limit = -1
		}
	}

	return
}

func (j *Jsonpath) evalArray(footprints []Footprint, node *ArrayNode) ([]Footprint, error) {
	if j.writeMode {
		for _, footprint := range footprints {
			tail := 0
			if !node.Params[0].Known {
				node.Params[0].Value = 0
			}
			if !node.Params[1].Known {
				tail = node.Params[0].Value + 1
			} else {
				tail = node.Params[1].Value
			}
			if node.Params[0].Value == 0 && node.Params[1].Value == 0 && node.Params[2].Value == 0 { // wildcard
				tail = -1
			}
			err := footprint.EnforceArraySelection(tail)
			if err != nil {
				return nil, err
			}
		}
	}
	footprints = expandFootprints(footprints, false)
	result := make([]Footprint, 0)
	for _, footprint := range footprints {
		ptr := footprint.HolderPtr()
		if arr, ok := (*ptr).([]interface{}); ok {
			base, limit, step, needInvert := j.inferArrayNode(&arr, node)
			indexes := make([]SelectionIndex, 0)
			realSize := footprint.(ArrayFootprint).RealSize
			if needInvert {
				for i := base; i < len(arr) && i > -1 && i > limit; i += step {
					indexes = append(indexes, SelectionIndex{
						Index: i,
						VirtualInfo: VirtualInfo{
							Virtual:  j.writeMode && i >= realSize,
							RealSize: -1,
						},
					})
				}
			} else {
				for i := base; i < len(arr) && i > -1 && i < limit; i += step {
					indexes = append(indexes, SelectionIndex{
						Index: i,
						VirtualInfo: VirtualInfo{
							Virtual:  j.writeMode && i >= realSize,
							RealSize: -1,
						},
					})
				}
			}
			result = append(result,
				ArrayFootprint{
					Ref:              footprint.HolderPtr(),
					SelectionIndexes: indexes,
				},
			)
		} else {
			j.AddWarning("cannot use a index number to find a element in a non-array object")
		}
	}
	return result, nil
}

func (j *Jsonpath) evalArrayElement(footprints []Footprint, node *ArrayElementNode) ([]Footprint, error) {
	if j.writeMode {
		if node.Value < 0 {
			return nil, fmt.Errorf("cannot use a negative index in set mode")
		} else if !node.Known {
			return nil, fmt.Errorf("index unknown in set mode")
		}
		for _, footprint := range footprints {
			err := footprint.EnforceArraySelection(node.Value + 1)
			if err != nil {
				return nil, err
			}
		}
	}
	footprints = expandFootprints(footprints, false)
	result := make([]Footprint, 0)
	for _, footprint := range footprints {
		ptr := footprint.HolderPtr()
		if arr, ok := (*ptr).([]interface{}); ok {
			indexes := make([]SelectionIndex, 0)
			realSize := footprint.(ArrayFootprint).RealSize
			i := -1
			if node.Value >= 0 && node.Value <= len(arr)-1 {
				i = node.Value
			} else if node.Value >= -len(arr) {
				i = node.Value + len(arr)
			}

			if i >= 0 && i < len(arr) {
				indexes = append(indexes, SelectionIndex{
					Index: i,
					VirtualInfo: VirtualInfo{
						Virtual:  j.writeMode && i >= realSize,
						RealSize: -1,
					},
				})
			}

			result = append(result,
				ArrayFootprint{
					Ref:              footprint.HolderPtr(),
					SelectionIndexes: indexes,
				},
			)
		} else {
			j.AddWarning("cannot use a index number to find a element in a non-array object")
		}
	}
	return result, nil
}

func (j *Jsonpath) evalWildcard(footprints []Footprint, node *WildcardNode) ([]Footprint, error) {
	footprints = expandFootprints(footprints, false)
	for i, footprint := range footprints {
		selected, err := footprint.SelectAll()
		if err != nil {
			log.Println("wildcard is only supported by map and array")
		} else {
			footprints[i] = selected
		}
	}
	return footprints, nil
}

func (j *Jsonpath) evalUnion(footprints []Footprint, node *UnionNode) ([]Footprint, error) {
	result := make([]Footprint, 0)
	for _, n := range node.Nodes {
		list, err := j.evalList(footprints, n)
		if err != nil {
			return footprints, err
		}
		result = append(result, list...)
	}
	return result, nil
}

func (j *Jsonpath) evalFilter(footprints []Footprint, node *FilterNode) ([]Footprint, error) {
	footprints = expandFootprints(footprints, false)
	result := make([]Footprint, 0)
	for _, fp := range footprints {
		allSelectedFp, err := fp.SelectAll()
		if err != nil {
			continue
		}
		elements, err := allSelectedFp.Expand()
		for _, element := range elements {
			element = element.LeaveItAsItIs()
			lefts, err := j.evalList([]Footprint{element}, node.Left)
			if node.Operator == "exists" {
				if len(lefts) > 0 {
					result = append(result, element)
				}
				continue
			}
			if err != nil {
				return nil, err
			}
			lefts = expandFootprints(lefts, true)

			var left, right interface{}
			switch {
			case len(lefts) == 0:
				continue
			case len(lefts) > 1:
				return nil, fmt.Errorf("can only compare one element at a time")
			}
			left = *(lefts[0].HolderPtr())

			rights, err := j.evalList([]Footprint{element}, node.Right)
			if err != nil {
				return nil, err
			}
			rights = expandFootprints(rights, true)
			switch {
			case len(rights) == 0:
				continue
			case len(rights) > 1:
				return nil, fmt.Errorf("can only compare one element at a time")
			}
			right = *(rights[0].HolderPtr())

			pass, err := genericCompare(node.Operator, left, right)
			if err != nil {
				j.AddWarning(err.Error())
			}
			if pass {
				result = append(result, element)
			}
		}
	}
	return result, nil
}

func genericCompare(operator string, left interface{}, right interface{}) (bool, error) {
	pass := false
	var err error
	switch operator {
	case "<":
		pass, err = template.Less(left, right)
	case ">":
		pass, err = template.Greater(left, right)
	case "==":
		pass, err = template.Equal(left, right)
	case "!=":
		pass, err = template.NotEqual(left, right)
	case "<=":
		pass, err = template.LessEqual(left, right)
	case ">=":
		pass, err = template.GreaterEqual(left, right)
	default:
		return false, fmt.Errorf("unrecognized filter operator %s", operator)
	}
	if err != nil {
		return false, err
	}
	return pass, nil
}

func (j *Jsonpath) evalRecursive(footprints []Footprint, node *RecursiveNode) ([]Footprint, error) {
	footprints = expandFootprints(footprints, false)
	result := make([]Footprint, 0)
	for _, footprint := range footprints {
		recursivelyCollectFootprint(footprint, &result)
	}
	return result, nil
}

func recursivelyCollectFootprint(footprint Footprint, result *[]Footprint) {
	*result = append(*result, footprint.LeaveItAsItIs()) // record self in result
	var err error
	if footprint, err = footprint.SelectAll(); err != nil {
		return
	}
	children, _ := footprint.Expand()
	for _, child := range children {
		recursivelyCollectFootprint(child, result)
	}
}

func (j *Jsonpath) evalInt(footprints []Footprint, node *IntNode) ([]Footprint, error) {
	footprints = expandFootprints(footprints, false)
	result := make([]Footprint, len(footprints))
	for i, _ := range footprints {
		var v interface{} = node.Value
		result[i] = NewFootprint(&v, nil)
	}
	return result, nil
}

func (j *Jsonpath) evalBool(footprints []Footprint, node *BoolNode) ([]Footprint, error) {
	footprints = expandFootprints(footprints, false)
	result := make([]Footprint, len(footprints))
	for i, _ := range footprints {
		var v interface{} = node.Value
		result[i] = NewFootprint(&v, nil)
	}
	return result, nil
}

func (j *Jsonpath) evalFloat(footprints []Footprint, node *FloatNode) ([]Footprint, error) {
	footprints = expandFootprints(footprints, false)
	result := make([]Footprint, len(footprints))
	for i, _ := range footprints {
		var v interface{} = node.Value
		result[i] = NewFootprint(&v, nil)
	}
	return result, nil
}
