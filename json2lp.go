package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type JSONinput struct {
	Results []Result `json:"results"`
	// Page   int      `json:"page"`
	// Fruits []string `json:"fruits"`
}

type Result struct {
	Series []Series `json:"series"`
}

type Series struct {
	Name    string          `json:"name"`
	Columns []string        `json:"columns"`
	Values  [][]interface{} `json:"values"`
}

type ProcessedJSON struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags"`
	Fields      map[string]interface{} `json:"fields"`
	Timestamp   time.Time              `json:"timestamp"`
}

// type Value struct {
// 	Data []interface{}
// }

func printColumnNames(data JSONinput) {
	for i, res := range data.Results {
		fmt.Printf("Result %d:\n", i)
		for j, ser := range res.Series {
			fmt.Printf("\tSeries %d (Table \"%s\"):\n", j, ser.Name)
			for k, col := range ser.Columns {
				fmt.Printf("\t\tColumn %d: %v\n", k, col)
			}
			for l, val := range ser.Values {
				fmt.Printf("\t\tValue %d:\n", l)
				for m, meas := range val {
					fmt.Println("\t\t\t", ser.Columns[m], "=", meas)
					// fmt.Printf("\t\t\tMeasurement %d: %+v\n",m, meas)
				}
			}
		}
	}
}

func ProcessJSON(input JSONinput, defs map[string]string) []ProcessedJSON {

	// map data to the right types for writing
	var output []ProcessedJSON
	for _, res := range input.Results {
		for _, ser := range res.Series {
			var timeIndex int
			tags := make(map[string]int)
			fields := make(map[string]int)
			// map definitions to columns
			for i, k := range ser.Columns {
				switch defs[k] {
				case "timestamp":
					timeIndex = i
				case "tag":
					tags[k] = i
				case "field":
					fields[k] = i
				case "ignore":
				}
			}
			for _, row := range ser.Values {
				value := ProcessedJSON{
					Measurement: "",
					Tags:        make(map[string]string),
					Fields:      make(map[string]interface{}),
				}

				//value.Measurement
				value.Measurement = ser.Name

				//value.Timestamp
				unixTime,_:=row[timeIndex].(float64)
				value.Timestamp = time.Unix(0,int64(unixTime)) 

				//value.Tags
				for k, i := range tags {
					value.Tags[k], _ = row[i].(string)
				}
				//value.Fields
				for k, i := range fields {
					value.Fields[k], _ = row[i].(float64)
				}
				output = append(output, value)
			}
		}
	}
	return output
}

// WriteOne writes a single JSON measurement to InfluxDB with
// an aync, non-blocking client you supply.
func WriteOne(writeAPI *api.WriteAPI, data ProcessedJSON, counter int) {
	client := *writeAPI

	p := influxdb2.NewPoint(
		data.Measurement,
		data.Tags,
		data.Fields,
		data.Timestamp)

	client.WritePoint(p)
	// Output a dot (.) for every successful write to influx
	// This helps people like me who need to see something to know it works
	if counter % 1000 == 0 {
		fmt.Printf("%d records uploaded @ %d:%d:%d, latest record: %+v\n", counter, time.Now().Hour(),time.Now().Minute(),time.Now().Second(), data)
	}
	//fmt.Printf(".")
}

// DumpToInflux loops through all the data you send it and writes all
// the points to Influx.
func DumpToInflux(data []ProcessedJSON) {
	influxIP, ok := os.LookupEnv("INFLUX_IP")
	if !ok {
		err := fmt.Errorf("INFLUX_IP not set.")
		fmt.Println(err)
		os.Exit(1)
	}
	client := influxdb2.NewClientWithOptions(fmt.Sprintf("http://%s:8086", influxIP), "my-token", influxdb2.DefaultOptions().SetBatchSize(20))
	writeAPI := client.WriteAPI("admin", "tesla")

	// The way this is set up, these likely don't get executed on ^C.
	defer client.Close()
	defer writeAPI.Flush()

	// Simple, isn't it?
	for i, point := range data {
		go WriteOne(&writeAPI, point, i)
		time.Sleep(time.Millisecond * 1)
	}
}

func main() {
	colorReset := "\033[0m"
	colorGreen := "\033[32m"
	colorRed := "\033[31m"

	args := os.Args[1:]

	if len(args) < 2 {
		fmt.Println(string(colorGreen), "json2lp v0.0.1", string(colorReset))
		fmt.Println("--------------")
		fmt.Println("\nConvert JSON as exported via Flux/InfluxDB to InfluxDB line protocol.")
		fmt.Println("\nSyntax: json2lp <json-file-name> <definitions-file-name>")
		os.Exit(0)
	}
	filename1 := args[0]

	file1, err := os.ReadFile(filename1)
	if err != nil {
		fmt.Println(string(colorRed), err, string(colorReset))
		os.Exit(1)
	}

	var data1 JSONinput

	if err := json.Unmarshal(file1, &data1); err != nil {
		panic(err)
	}

	filename2 := args[1]

	file2, err := os.ReadFile(filename2)
	if err != nil {
		fmt.Println(string(colorRed), err, string(colorReset))
		os.Exit(1)
	}

	var data2 map[string]string

	if err := json.Unmarshal(file2, &data2); err != nil {
		panic(err)
	}

	output := ProcessJSON(data1, data2)
	DumpToInflux(output)
	//fmt.Printf("%+v", output)
}
