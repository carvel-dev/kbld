// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Path []*PathPart

type PathPart struct {
	MapKey     *string             `json:"mapKey,omitempty"`
	ArrayIndex *PathPartArrayIndex `json:"arrayIndex,omitempty"`
}

var _ json.Unmarshaler = &PathPart{}
var _ json.Marshaler = Path{}

type PathPartArrayIndex struct {
	Index *int  `json:"index,omitempty"`
	All   *bool `json:"allIndexes,omitempty"`
}

func NewPathFromStrings(strs []string) Path {
	var path Path
	for _, str := range strs {
		path = append(path, NewPathPartFromString(str))
	}
	return path
}

func NewPathFromInterfaces(parts []interface{}) Path {
	var path Path
	for _, part := range parts {
		switch typedPart := part.(type) {
		case string:
			path = append(path, NewPathPartFromString(typedPart))
		case int:
			path = append(path, NewPathPartFromIndex(typedPart))
		default:
			panic(fmt.Sprintf("Unexpected part type %T", typedPart))
		}
	}
	return path
}

func (p Path) AsStrings() []string {
	var result []string
	for _, part := range p {
		if part.MapKey == nil {
			panic(fmt.Sprintf("Unexpected non-map-key path part '%#v'", part))
		}
		result = append(result, *part.MapKey)
	}
	return result
}

func (p Path) AsString() string {
	var result []string
	for _, part := range p {
		result = append(result, part.AsString())
	}
	return strings.Join(result, ",")
}

func (p Path) ContainsNonMapKeys() bool {
	for _, part := range p {
		if part.MapKey == nil {
			return true
		}
	}
	return false
}

func (p Path) MarshalJSON() ([]byte, error) {
	var result []interface{}
	for _, part := range p {
		switch {
		case part.MapKey != nil:
			result = append(result, *part.MapKey)
		case part.ArrayIndex != nil:
			result = append(result, part.ArrayIndex)
		default:
			panic("Unknown path part")
		}
	}
	return json.Marshal(result)
}

func NewPathPartFromString(str string) *PathPart {
	return &PathPart{MapKey: &str}
}

func NewPathPartFromIndex(i int) *PathPart {
	return &PathPart{ArrayIndex: &PathPartArrayIndex{Index: &i}}
}

func NewPathPartFromIndexAll() *PathPart {
	trueBool := true
	return &PathPart{ArrayIndex: &PathPartArrayIndex{All: &trueBool}}
}

func (p *PathPart) AsString() string {
	switch {
	case p.MapKey != nil:
		return *p.MapKey
	case p.ArrayIndex != nil && p.ArrayIndex.Index != nil:
		return fmt.Sprintf("%d", *p.ArrayIndex.Index)
	case p.ArrayIndex != nil && p.ArrayIndex.All != nil:
		return "(all)"
	default:
		panic("Unknown path part")
	}
}

func (p *PathPart) UnmarshalJSON(data []byte) error {
	var str string
	var idx PathPartArrayIndex

	switch {
	case json.Unmarshal(data, &str) == nil:
		p.MapKey = &str
	case json.Unmarshal(data, &idx) == nil:
		p.ArrayIndex = &idx
	default:
		return fmt.Errorf("Unknown path part")
	}
	return nil
}

func (p Path) Matches(p2 Path) bool {
	if len(p) != len(p2) {
		return false
	}
	for i := range p {
		if !p[i].Matches(p2[i]) {
			return false
		}
	}
	return true
}

func (p Path) HasMatchingSuffix(ending Path) bool {
	if len(p) < len(ending) {
		return false
	}
	for i := range ending {
		if !p[len(p)-1-i].Matches(ending[len(ending)-1-i]) {
			return false
		}
	}
	return true
}

func (p *PathPart) Matches(p2 *PathPart) bool {
	switch {
	case p.MapKey != nil && p2.MapKey != nil:
		return *p.MapKey == *p2.MapKey

	case p.ArrayIndex != nil && p2.ArrayIndex != nil:
		switch {
		case p.ArrayIndex.Index != nil && p2.ArrayIndex.Index != nil:
			return *p.ArrayIndex.Index == *p2.ArrayIndex.Index
		case p.ArrayIndex.All != nil && p2.ArrayIndex.Index != nil:
			return true
		default:
			return false
		}

	default:
		return false
	}
}
