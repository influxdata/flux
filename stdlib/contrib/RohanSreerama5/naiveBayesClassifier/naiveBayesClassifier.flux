//Final working code as of August 11, 2020
//Currently supports single field classification and binary data sets 
//Please ensure Ruby is installed

package naiveBayesClassifier

import "system"

naiveBayes = (tables=<-, myClass, myField, myMeasurement) => {
 
training_data = tables
  |> range(start:2020-01-02T00:00:00Z, stop: 2020-01-06T23:00:00Z) //data for 3 days 
  |> filter(fn: (r) => r["_measurement"] == myMeasurement and r["_field"] == myField)
  |> group()
  //|> yield(name: "trainingData")

test_data = tables
  |> range(start:2020-01-01T00:00:00Z, stop: 2020-01-01T23:00:00Z) //data for 1 day
  |> filter(fn: (r) => r["_measurement"] == myMeasurement and r["_field"] == myField)
  |> group()
  //|> yield(name: "test data")

//data preparation 
r = training_data
	|> group(columns: ["_field"])
	|> count()
	|> tableFind(fn: (key) =>
		(key._field == myField))
r2 = getRecord(table: r, idx: 0)
total_count = r2._value
P_Class_k = training_data
	|> group(columns: [myClass, "_field"])
	|> count()
	|> map(fn: (r) =>
		({r with p_k: float(v: r._value) / float(v: total_count), tc: total_count}))
	|> group()

//one table for each class, where r.p_k == P(Class_k)
P_value_x = training_data
	|> group(columns: ["_value", "_field"])
	|> count(column: myClass)
	|> map(fn: (r) =>
		({r with p_x: float(v: r.airborne) / float(v: total_count), tc: total_count}))

// one table for each value, where r.p_x == P(value_x)
P_k_x = training_data
	|> group(columns: ["_field","_value", myClass])
	|> reduce(fn: (r, accumulator) =>
		({sum: 1.0 + accumulator.sum}), identity: {sum: 0.0})
	|> group()

// one table for each value and Class pair, where r.p_k_x == P(value_x | Class_k)
P_k_x_class = join(tables: {P_k_x: P_k_x, P_Class_k: P_Class_k}, on: [myClass], method: "inner")
    |> group(columns: [myClass, "_value_P_k_x"])
	  |> limit(n: 1)

	  |> map(fn: (r) =>
		  ({r with P_x_k: r.sum / float(v: r._value_P_Class_k)}))
    |> drop(columns: ["_field_P_Class_k", "_value_P_Class_k"])
    |> rename(columns: {_field_P_k_x: "_field", _value_P_k_x: "_value"})

P_k_x_class_Drop = join(tables: {P_k_x: P_k_x, P_Class_k: P_Class_k}, on: [myClass], method: "inner")
    |> drop(columns: ["_field_P_Class_k", "_value_P_Class_k", "_field_P_k_x"])
    |> group(columns: [myClass, "_value_P_k_x"])
	|> limit(n: 1)

	|> map(fn: (r) =>
		({r with P_x_k: r.sum / float(v: r._value_P_Class_k)}))

//added P(value_x) to table
//calculated probabilities for training data 
Probability_table = join(tables: {P_k_x_class: P_k_x_class, P_value_x: P_value_x}, on: ["_value", "_field"], method: "inner")
	|> map(fn: (r) =>
		({r with Probability: r.P_x_k * r.p_k / r.p_x}))
	
 	//|> yield(name: "final")

//predictions for test data computed 
predictOverall = (tables=<-) => {
  r = tables
    |> keep(columns: ["_value", "Animal_name","_field"])

  output = join(tables: {Probability_table: Probability_table, r: r}, on: ["_value"], method: "inner")
  return output 
}

return test_data |> predictOverall() 

}