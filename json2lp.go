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
	Measurement string `json:"measurement"`
	Tags map[string]string `json:"tags"`
	Fields map[string]interface{} `json:"fields"`
	Timestamp time.Time `json:"timestamp"`
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

// WriteOne writes a single JSON measurement to InfluxDB with
// an aync, non-blocking client you supply.
func WriteOne(writeAPI *api.WriteAPI, data ProcessedJSON) {
	client := *writeAPI

	p := influxdb2.NewPoint(
		data.Measurement,
		data.Tags,
		data.Fields,
		data.Timestamp)
		
	client.WritePoint(p)
	// Output a dot (.) for every successful write to influx
	// This helps people like me who need to see something to know it works
	fmt.Printf(".")
}

// DumpToInflux simply runs the totality of polly in your program. It is
// recommended you run this as a goroutine so your program can do
// other things.
func DumpToInflux(data []ProcessedJSON) {
	influxIP, ok := os.LookupEnv("INFLUX_IP")
	if !ok {
		err := fmt.Errorf("INFLUX_IP not set.")
		fmt.Println(err)
		os.Exit(1)
	}
	client := influxdb2.NewClientWithOptions(fmt.Sprintf("http://%s:8086", influxIP), "my-token", influxdb2.DefaultOptions().SetBatchSize(20))
	writeAPI := client.WriteAPI("myorg", "hpwc")

	// The way this is set up, these likely don't get executed on ^C.
	defer client.Close()
	defer writeAPI.Flush()

	// Simple, isn't it?
	for _, point := range data {
		go WriteOne(&writeAPI, point)
		time.Sleep(time.Millisecond * 10)
	}
}

func main() {
	colorReset := "\033[0m"
	colorGreen := "\033[32m"
	colorRed := "\033[31m"

	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println(string(colorGreen), "json2lp v0.0.1", string(colorReset))
		fmt.Println("--------------")
		fmt.Println("\nConvert JSON as exported via Flux/InfluxDB to InfluxDB line protocol.")
		fmt.Println("\nSyntax: json2lp <json-file-name>")
		os.Exit(0)
	}
	filename := args[0]

	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(string(colorRed), err, string(colorReset))
		os.Exit(1)
	}

	var data JSONinput

	if err := json.Unmarshal(file, &data); err != nil {
		panic(err)
	}

	fmt.Printf("%sUseful details about JSON input:%s\n", colorGreen, colorReset)
	fmt.Printf("%s\tActual Data:%s %+v\n", colorGreen, colorReset, data)
	printColumnNames(data)
}
