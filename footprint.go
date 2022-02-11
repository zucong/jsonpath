package jsonpath

import (
	"errors"
	"fmt"
)

type Footprint interface {
	LeaveItAsItIs() Footprint
	Expand() ([]Footprint, error)
	HolderPtr() *interface{}
	UpdateOne(data interface{}, keyOrIndex interface{}) error
	UpdateAll(data interface{}) error
	SelectAll() (Footprint, error)
	IsVirtual() bool
	EnforceArraySelection(size int) error
	EnforceObjectSelection() error
}

type VirtualInfo struct {
	Virtual  bool
	RealSize int
}

type SelectionKey struct {
	Key string
	VirtualInfo
}

type MapFootprint struct {
	leaveItAsItIs bool
	Ref           *interface{}
	SelectionKeys []SelectionKey
	Virtual       bool
}

func NewFootprint(ptr *interface{}, virtualInfo interface{}) Footprint {
	var virtual bool
	var realSize int
	if sk, ok := virtualInfo.(SelectionKey); ok {
		virtual = sk.Virtual
		realSize = sk.RealSize
	} else if si, ok := virtualInfo.(SelectionIndex); ok {
		virtual = si.Virtual
		realSize = si.RealSize
	}

	if _, ok := (*ptr).(map[string]interface{}); ok {
		return MapFootprint{
			Ref:           ptr,
			SelectionKeys: nil,
			Virtual:       virtual,
		}
	} else if _, ok := (*ptr).([]interface{}); ok {
		return ArrayFootprint{
			Ref:              ptr,
			SelectionIndexes: nil,
			VirtualInfo: VirtualInfo{
				Virtual:  virtual,
				RealSize: realSize,
			},
		}
	} else {
		return NonRefFootprint{
			value: *ptr,
		}
	}
}

func (mfp MapFootprint) LeaveItAsItIs() Footprint {
	mfp.leaveItAsItIs = true
	return mfp
}

func (mfp MapFootprint) Expand() ([]Footprint, error) {
	if mfp.leaveItAsItIs {
		mfp.leaveItAsItIs = false
		return []Footprint{mfp}, nil
	}
	if len(mfp.SelectionKeys) == 0 {
		return nil, nil
	}
	result := make([]Footprint, 0)
	ref := (*mfp.Ref).(map[string]interface{})
	for _, sk := range mfp.SelectionKeys {
		v := ref[sk.Key]
		result = append(result, NewFootprint(&v, sk))
	}
	return result, nil
}

func (mfp MapFootprint) HolderPtr() *interface{} {
	return mfp.Ref
}

func (mfp MapFootprint) UpdateAll(data interface{}) error {
	ref := (*mfp.Ref).(map[string]interface{})
	for _, sk := range mfp.SelectionKeys {
		ref[sk.Key] = data
	}
	return nil
}

func (mfp MapFootprint) UpdateOne(data interface{}, keyOrIndex interface{}) error {
	if key, ok := keyOrIndex.(string); ok {
		(*mfp.Ref).(map[string]interface{})[key] = data
	} else {
		return errors.New("cannot extract key")
	}
	return nil
}

func (mfp MapFootprint) SelectAll() (Footprint, error) {
	ref := (*mfp.Ref).(map[string]interface{})
	sks := make([]SelectionKey, 0)
	for key := range ref {
		sks = append(sks, SelectionKey{
			Key: key,
			VirtualInfo: VirtualInfo{
				Virtual:  false,
				RealSize: -1,
			},
		})
	}
	mfp.SelectionKeys = sks
	return mfp, nil
}

func (mfp MapFootprint) EnforceArraySelection(size int) error {
	ref := (*mfp.Ref).(map[string]interface{})
	for i, s := range mfp.SelectionKeys {
		if _, ok := ref[s.Key]; !ok {
			return fmt.Errorf("cannot find the element by key: %s", s.Key)
		}

		if _, ok := ref[s.Key].([]interface{}); ok {
			s.RealSize = len(ref[s.Key].([]interface{}))
			if size != -1 && s.RealSize < size {
				arr := make([]interface{}, size-s.RealSize)
				ref[s.Key] = append(ref[s.Key].([]interface{}), arr...)
			}
		} else if !ok {
			if !s.Virtual {
				return fmt.Errorf("the selection is not an array or a virtual")
			}
			if size == -1 {
				return fmt.Errorf("cannot use * to set in a virtual")
			}
			s.RealSize = -1
			ref[s.Key] = make([]interface{}, size)
		}
		mfp.SelectionKeys[i] = s
	}
	return nil
}

func (mfp MapFootprint) EnforceObjectSelection() error {
	ref := (*mfp.Ref).(map[string]interface{})
	for _, s := range mfp.SelectionKeys {
		if _, ok := ref[s.Key]; !ok {
			return fmt.Errorf("cannot find the element by key: %s", s.Key)
		}
		if _, ok := ref[s.Key].(map[string]interface{}); !ok {
			if s.Virtual {
				ref[s.Key] = make(map[string]interface{}, 0)
			} else {
				return fmt.Errorf("the selection is not an array or a virtual")
			}
		}
	}
	return nil
}

func (mfp MapFootprint) IsVirtual() bool {
	return mfp.Virtual
}

type SelectionIndex struct {
	Index int
	VirtualInfo
}

type ArrayFootprint struct {
	leaveItAsItIs    bool
	Ref              *interface{}
	SelectionIndexes []SelectionIndex
	VirtualInfo
}

func (afp ArrayFootprint) LeaveItAsItIs() Footprint {
	afp.leaveItAsItIs = true
	return afp
}

func (afp ArrayFootprint) Expand() ([]Footprint, error) {
	if afp.leaveItAsItIs {
		afp.leaveItAsItIs = false
		return []Footprint{afp}, nil
	}
	if len(afp.SelectionIndexes) == 0 {
		return nil, nil
	}
	result := make([]Footprint, 0)
	ref := (*afp.Ref).([]interface{})
	for _, s := range afp.SelectionIndexes {
		v := ref[s.Index]

		result = append(result, NewFootprint(&v, s))
	}
	return result, nil
}

func (afp ArrayFootprint) HolderPtr() *interface{} {
	return afp.Ref
}

func (afp ArrayFootprint) UpdateAll(data interface{}) error {
	ref := (*afp.Ref).([]interface{})
	for _, si := range afp.SelectionIndexes {
		ref[si.Index] = data
	}
	return nil
}

func (afp ArrayFootprint) UpdateOne(data interface{}, keyOrIndex interface{}) error {
	if key, ok := keyOrIndex.(int); ok {
		(*afp.Ref).([]interface{})[key] = data
	} else {
		return errors.New("cannot extract index")
	}
	return nil
}

func (afp ArrayFootprint) SelectAll() (Footprint, error) {
	ref := (*afp.Ref).([]interface{})
	selection := make([]SelectionIndex, len(ref))
	for i := 0; i < len(ref); i++ {
		selection[i] = SelectionIndex{
			Index: i,
			VirtualInfo: VirtualInfo{
				Virtual:  false,
				RealSize: -1,
			},
		}
	}
	afp.SelectionIndexes = selection
	return afp, nil
}

func (afp ArrayFootprint) IsVirtual() bool {
	return afp.Virtual
}

func (afp ArrayFootprint) EnforceArraySelection(size int) error {
	ref := (*afp.Ref).([]interface{})
	for i, s := range afp.SelectionIndexes {
		if s.Index < 0 || s.Index > len(ref) {
			return fmt.Errorf("invalid index when EnforceArraySelection: %d", s.Index)
		}

		if _, ok := ref[s.Index].([]interface{}); ok {
			s.RealSize = len(ref[s.Index].([]interface{}))
			if size != -1 && s.RealSize < size {
				arr := make([]interface{}, size-s.RealSize)
				ref[s.Index] = append(ref[s.Index].([]interface{}), arr...)
			}
		} else if !ok {
			if !s.Virtual {
				return fmt.Errorf("the selection is not an array or a virtual")
			}
			if size == -1 {
				return fmt.Errorf("cannot use * to set in a virtual")
			}
			s.RealSize = -1
			ref[s.Index] = make([]interface{}, size)
		}
		afp.SelectionIndexes[i] = s
	}
	return nil
}

func (afp ArrayFootprint) EnforceObjectSelection() error {
	ref := (*afp.Ref).([]interface{})
	for _, s := range afp.SelectionIndexes {
		if s.Index < 0 || s.Index > len(ref) {
			return fmt.Errorf("invalid index when EnforceObjectSelection: %d", s.Index)
		}
		if _, ok := ref[s.Index].(map[string]interface{}); !ok {
			if s.Virtual {
				ref[s.Index] = make(map[string]interface{}, 0)
			} else {
				return fmt.Errorf("the selection is not an array or a virtual")
			}
		}
	}
	return nil
}

type NonRefFootprint struct {
	leaveItAsItIs bool
	value         interface{}
}

func (nfp NonRefFootprint) LeaveItAsItIs() Footprint {
	nfp.leaveItAsItIs = true
	return nfp
}

func (nfp NonRefFootprint) Expand() ([]Footprint, error) {
	nfp.leaveItAsItIs = false
	if nfp.leaveItAsItIs {
		return []Footprint{nfp}, nil
	}
	return nil, errors.New("non-reference foot print cannot be expand")
}

func (nfp NonRefFootprint) HolderPtr() *interface{} {
	return &nfp.value
}

func (nfp NonRefFootprint) UpdateAll(data interface{}) error {
	return errors.New("UpdateAll is not supported by NonRefFootprint")
}

func (nfp NonRefFootprint) UpdateOne(data interface{}, keyOrIndex interface{}) error {
	return errors.New("UpdateOne is not supported by NonRefFootprint")
}

func (nfp NonRefFootprint) SelectAll() (Footprint, error) {
	return nil, errors.New("SelectAll is not supported by NonRefFootprint")
}

func (nfp NonRefFootprint) IsVirtual() bool {
	return false
}

func (nfp NonRefFootprint) EnforceArraySelection(size int) error {
	return fmt.Errorf("EnforceArraySelection is not supported by NonRefFootprint")
}

func (nfp NonRefFootprint) EnforceObjectSelection() error {
	return fmt.Errorf("EnforceObjectSelection is not supported by NonRefFootprint")
}
