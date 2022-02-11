# json2lp
Go command line tool for converting json exported data from InfluxDB to line protocol. Helpful for enabling backup and restore operations in poorly-designed InfluxDB setups.

Reads JSON outputs from Influx 1.x queries, parses this data into individual data points, and loads them into a target InfluxDB instance. A cheap way to get around forgetting to enable backups!

## How To Use This
1. Run `make build` to build the binary for your system
2. Supply a JSON output of an Influx query (e.g. something like `influx -host 192.168.86.32 -database tesla -format json -execute "select * from utilities" > utilities-backup-2022-01-24.json`) This is a somewhat tricky step because this particular incantation requires InfluxDB 1.x (1.9.6 works fine)
3. Modify `definitions.json` to set rules for the columns in the output of your query. The keys should match the column names from the output, and the values are what you set. Your options are `timestamp` (only one column is allowed!), `tag`, `field`, or `ignore`.
4. Supply an IP address for your Influx server by exporting it as an env var (e.g. `export INFLUX_IP="192.168.1.23"`)
5. Run the executable:
```bash
./json2lp my-big-backup.json definitions.json
```
The first argument is always the data you want to load, and the second argument is the column mapping.

## Example Data
This is a sampling of what actual result data from an influx query looks like in JSON format:
```json
{
  "results": [
    {
      "series": [
        {
          "name": "utilities",
          "columns": [
            "time",
            "consumption",
            "endpoint_id",
            "endpoint_type",
            "interval",
            "msg_type",
            "outage",
            "protocol"
          ],
          "values": [
            [
              1614894082033813000,
              50355,
              "5678",
              "4",
              null,
              "cumulative",
              null,
              "SCM"
            ],
            [
              1614894082367864000,
              28989,
              "12345",
              "4",
              null,
              "cumulative",
              null,
              "SCM"
            ],
            [
              1614894082589024000,
              28989,
              "12345",
              "4",
              null,
              "cumulative",
              null,
              "SCM"
            ],
            [
              1643059222041570000,
              1978543,
              "9876",
              "110",
              null,
              "cumulative",
              null,
              "SCM+"
            ]
          ]
        }
      ]
    }
  ]
}
```

This represents 4 data points in a single table within a series called "utilites". `json2lp` will extract the series name(s) but is currently only capable of dealing with a single schema - it will apply your `definitions.json` mappings to all series.

## To-Do List
[] Speed up Influx writes using buffering/batching/whatever (go as fast as the server can take it)
[] Try outputting the 10MB max of actual line protocol as another way to pour data in
[] Tighten up the interface (guardrails, self-descriptive instructions)
[] Tighten up the docs
