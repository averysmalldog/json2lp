package main

import (
	"encoding/json"
	"fmt"
	"os"
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
	Name string `json:"name"`
	Columns []string `json:"columns"`
	Values []interface{} `json:"values"`
}

// type Value struct {
// 	Data []interface{}
// }

func main() {
	colorReset := "\033[0m"
	colorGreen := "\033[32m"

	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println(string(colorGreen),"json2lp v0.0.1",string(colorReset))
		fmt.Println("--------------")
		fmt.Println("\nConvert JSON as exported via Flux/InfluxDB to InfluxDB line protocol.")
		fmt.Println("\nSyntax: json2lp <json-file-name>")
		os.Exit(0)
	}
	filename := args[0]

	file, err := os.ReadFile(filename)
	if err != nil {
		os.Exit(1)
	}

	var data JSONinput

	if err := json.Unmarshal(file, &data); err != nil {
        panic(err)
    }
    // fmt.Println(data)
	fmt.Println(len(data.Results[0].Series[0].Values))
	// file, err := os.Open(filename)
	// if err != nil {
	// 	fmt.Println("Something went wrong opening the file:", err)
	// 	os.Exit(1)
	// } else {
	// 	defer file.Close()
	// 	reader := csv2lp.CsvToLineProtocol(file)
	// 	buffer := make([]byte, 1024)
	// 	for {
	// 		n, e := reader.Read(buffer)
	// 		if e != nil {
	// 			if e == io.EOF {
	// 				break
	// 			} else {
	// 				fmt.Println("Error:", e)
	// 				break
	// 			}
	// 		}
	// 		fmt.Print(string(buffer[:n]))
	// 	}
	// }
}
