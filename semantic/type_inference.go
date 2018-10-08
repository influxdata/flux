package semantic

func InferTypes(n Node) (SolutionMap, error) {
	annotations := Annotate(n)
	constraints := GenerateConstraints(n, annotations)
	return SolveTypes(annotations, constraints)
}
