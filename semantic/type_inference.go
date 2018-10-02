package semantic

func InferTypes(program *Program) (SolutionMap, error) {
	annotations := Annotate(program)
	constraints := GenerateConstraints(program, annotations)
	return SolveTypes(program, annotations, constraints)
}
