package math

builtin minIndex : (values: [A]) => int where A: Numeric
min = (values) => {
    index = minIndex(values)
    return values[index]
}

builtin maxIndex : (values: [A]) => int where A: Numeric
max = (values) => {
	index = maxIndex(values)
	return values[index]
}

builtin sum : (values: [A]) => A where A: Numeric