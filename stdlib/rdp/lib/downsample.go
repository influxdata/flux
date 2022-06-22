package rdp

//import "fmt"

func DownSampleIpdata(ipdata []Point, threshold float64) []Point {
	if len(ipdata) <= 2 {
		return ipdata
	}
	line := ConnectLine{Starting: ipdata[0], Ending: ipdata[len(ipdata)-1]}

	farthest_point_index, farthest_point_distance_from_line := FindFarthestPoint(line, ipdata)
	if farthest_point_distance_from_line >= threshold {
		left := DownSampleIpdata(ipdata[:farthest_point_index+1], threshold)
		right := DownSampleIpdata(ipdata[farthest_point_index:], threshold)
		return append(left[:len(left)-1], right...)
	}

	return []Point{ipdata[0], ipdata[len(ipdata)-1]}

}

func DownSampleIpdata_iter(ipdata []Point, start_index int64, last_index int64, threshold float64) []bool {
	var stack [][]int64
	global_start_index := start_index
	stack = append(stack, []int64{start_index, last_index})
	indices := make([]bool, last_index-start_index+1)
	for i, _ := range indices {
		indices[i] = true
	}
	for len(stack) > 0 {
		n := len(stack) - 1
		curr := stack[n]
		//fmt.Println(curr)
		stack = stack[:n]
		start_index := curr[0]
		last_index := curr[1]
		dmax := 0.0
		index := start_index
		line := ConnectLine{Starting: ipdata[start_index], Ending: ipdata[last_index]}
		for i := index + 1; i < last_index; i++ {
			if indices[i-global_start_index] {

				distance := line.perpendicularDistanceFromPointToLine(ipdata[i])
				if distance > dmax {
					index = i
					dmax = distance
				}

			}
		}
		if dmax > threshold {
			stack = append(stack, []int64{start_index, index})
			stack = append(stack, []int64{index, last_index})
		} else {
			for i := start_index + 1; i < last_index; i++ {
				indices[i-global_start_index] = false
			}
		}
	}
	return indices
}

func FindFarthestPoint(line ConnectLine, ipdata []Point) (farthest_point_index int, farthest_point_distance_from_line float64) {
	for i := 0; i < len(ipdata); i++ {
		distance := line.perpendicularDistanceFromPointToLine(ipdata[i])
		if distance > farthest_point_distance_from_line {
			farthest_point_distance_from_line = distance
			farthest_point_index = i
		}
	}

	return farthest_point_index, farthest_point_distance_from_line
}
