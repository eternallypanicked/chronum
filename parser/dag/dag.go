package dag

import "fmt"

type Node struct {
	ID           string
	Dependencies []*Node
	Dependents	 []*Node
}

type DAG struct {
	Nodes map[string]*Node
}

func (d *DAG) TopologicalSort() ([]*Node, error) {
    visited := make(map[string]bool)
    temp := make(map[string]bool)
    var result []*Node

    var visit func(*Node) error
    visit = func(n *Node) error {
        if temp[n.ID] {
            return fmt.Errorf("cycle detected at %s", n.ID)
        }
        if visited[n.ID] {
            return nil
        }
        temp[n.ID] = true
        for _, dep := range n.Dependencies {
            if err := visit(dep); err != nil {
                return err
            }
        }
        temp[n.ID] = false
        visited[n.ID] = true
        result = append(result, n)
        return nil
    }

    for _, n := range d.Nodes {
        if !visited[n.ID] {
            if err := visit(n); err != nil {
                return nil, err
            }
        }
    }

    return result, nil
}


func New() *DAG {
	return &DAG{Nodes: make(map[string]*Node)}
}

func (d *DAG) AddNode(id string) *Node {
	node := &Node{ID: id}
	d.Nodes[id] = node
	return node
}

func (d *DAG) LinkDependencies(steps map[string][]string) error {
	for id, deps := range steps {
		node := d.Nodes[id]
		for _, dep := range deps {
			target, ok := d.Nodes[dep]
			node.Dependencies = append(node.Dependencies, target)
			target.Dependents = append(target.Dependents, node)
			if !ok {
				return fmt.Errorf("unknown dependency: %s", dep)
			}
			node.Dependencies = append(node.Dependencies, target)
		}
	}
	return nil
}
