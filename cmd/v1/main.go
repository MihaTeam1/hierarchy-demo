package main

import (
	"fmt"
	"slices"
)

func main() {
	s := NewStorage()
	p := NewPolicyManager(PolicyManagerDependency{
		EmployeeSource:  s,
		PositionSource:  s,
		AttributeSource: s,
	})

	emp := s.ListEmployees(nil)[1]
	managers, err := p.GetFirstManagers(emp.ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, manager := range managers {
		fmt.Println(manager.ID, manager.Name, manager.PositionID)
	}
}

type Employee struct {
	ID         string
	Name       string
	PositionID string
}

type Position struct {
	ID   string
	Name string
}

type HierarchyAttributes struct {
	ID         string
	Type       string
	Attributes map[string]any
}

func GetAttribute[T any](a HierarchyAttributes, name string) (T, bool) {
	var resp T

	v, ok := a.Attributes[name]
	if !ok {
		return resp, false
	}

	resp, ok = v.(T)
	return resp, ok
}

type Storage struct {
	attributes []HierarchyAttributes
	employees  []Employee
	positions  []Position
}

func NewStorage() Storage {
	return Storage{
		attributes: []HierarchyAttributes{
			{
				ID:         "1",
				Type:       "employee",
				Attributes: map[string]any{},
			},
			{
				ID:   "2",
				Type: "position",
				Attributes: map[string]any{
					"managerPosition": "1",
				},
			},
		},
		positions: []Position{
			{
				ID:   "1",
				Name: "Мастер",
			},
			{
				ID:   "2",
				Name: "Пекарь",
			},
		},
		employees: []Employee{
			{
				ID:         "1",
				Name:       "John 1",
				PositionID: "1",
			},
			{
				ID:         "2",
				Name:       "John 2",
				PositionID: "2",
			},
			{
				ID:         "3",
				Name:       "John 3",
				PositionID: "1",
			},
		},
	}
}

func (s Storage) ListAttributes(t string) []HierarchyAttributes {
	if t == "" {
		return s.attributes
	}

	var resp []HierarchyAttributes
	for _, attr := range s.attributes {
		if attr.Type == t {
			resp = append(resp, attr)
		}
	}
	return resp
}

func (s Storage) ListEmployees(ids []string) []Employee {
	if ids == nil {
		return s.employees
	}

	var resp []Employee
	for _, emp := range s.employees {
		if slices.Contains(ids, emp.ID) {
			resp = append(resp, emp)
		}
	}
	return resp
}

func (s Storage) ListEmployeesByPosition(ids []string) []Employee {
	if ids == nil {
		return s.employees
	}

	var resp []Employee
	for _, emp := range s.employees {
		if slices.Contains(ids, emp.PositionID) {
			resp = append(resp, emp)
		}
	}
	return resp
}

func (s Storage) ListPositions(ids []string) []Position {
	if ids == nil {
		return s.positions
	}

	var resp []Position
	for _, pos := range s.positions {
		if slices.Contains(ids, pos.ID) {
			resp = append(resp, pos)
		}
	}

	return resp
}

type EmployeeSource interface {
	ListEmployees(ids []string) []Employee
	ListEmployeesByPosition(ids []string) []Employee
}

type PositionSource interface {
	ListPositions(ids []string) []Position
}

type AttributeSource interface {
	ListAttributes(t string) []HierarchyAttributes
}

type PolicyManager struct {
	d PolicyManagerDependency
}

type PolicyManagerDependency struct {
	EmployeeSource  EmployeeSource
	PositionSource  PositionSource
	AttributeSource AttributeSource
}

func NewPolicyManager(d PolicyManagerDependency) *PolicyManager {
	return &PolicyManager{
		d: d,
	}
}

func (p *PolicyManager) GetFirstManagers(empID string) ([]Employee, error) {
	var emp Employee
	if emps := p.d.EmployeeSource.ListEmployees([]string{empID}); len(emps) > 1 {
		return nil, fmt.Errorf("invalid emps resp")
	} else if len(emps) == 0 {
		return nil, fmt.Errorf("emp not found")
	} else {
		emp = emps[0]
	}

	var position Position
	if pos := p.d.PositionSource.ListPositions([]string{emp.PositionID}); len(pos) > 1 {
		return nil, fmt.Errorf("invalid pos resp")
	} else if len(pos) == 0 {
		return nil, fmt.Errorf("pos not found")
	} else {
		position = pos[0]
	}

	var attributes HierarchyAttributes
	attrs := p.d.AttributeSource.ListAttributes("position")
	for _, attr := range attrs {
		if attr.ID == position.ID {
			attributes = attr
		}
	}

	managerPositionID, ok := GetAttribute[string](attributes, "managerPosition")
	if !ok {
		return nil, nil // if there is no emp positions above it's ok
	}

	return p.d.EmployeeSource.ListEmployeesByPosition([]string{managerPositionID}), nil
}
