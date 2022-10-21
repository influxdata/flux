package plan

func cloneSpec(s *Spec) (*Spec, error) {
	nodeMap := make(map[Node]Node)
	s.TopDownWalk(func(node Node) error {
		nodeMap[node] = node.ShallowCopy()
		return nil
	})

	for oldNode, newNode := range nodeMap {
		newNode.ClearPredecessors()
		for _, n := range oldNode.Predecessors() {
			newNode.AddPredecessors(nodeMap[n])
		}

		newNode.ClearSuccessors()
		for _, n := range oldNode.Successors() {
			newNode.AddSuccessors(nodeMap[n])
		}
	}

	newRoots := make(map[Node]struct{})
	for r := range s.Roots {
		newRoots[nodeMap[r]] = struct{}{}
	}

	newSpec := &Spec{
		Roots:     newRoots,
		Resources: s.Resources,
		Now:       s.Now,
	}

	return newSpec, nil
}
