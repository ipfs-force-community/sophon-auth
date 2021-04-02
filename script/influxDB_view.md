### histogram
```
from(bucket: "bkt2")
  |> range(start: -1h)
  |> filter(fn: (r) => r["_measurement"] == "verify")
  |> filter(fn: (r) => r["_field"] == "method")
  |> aggregateWindow(every: 10s, fn: count, createEmpty: false)
  |> yield(name: "mean")

from(bucket: "bkt2")
  |> range(start: -1h)
  |> filter(fn: (r) => r["_measurement"] == "verify")
  |> filter(fn: (r) => r["_field"] == "method")
  |> sort()
  |> yield(name: "sort")

from(bucket: "bkt2")
  |> range(start:-1h)
  |> filter(fn: (r) => r["_measurement"] == "verify")
  |> filter(fn: (r) => r["_field"] == "method")
  |> count()
  |> yield(name: "count")
  
```

### graph
```
from(bucket: "bkt2")
  |> range(start: v.timeRangeStart, stop: v.timeRangeStop)
  |> filter(fn: (r) => r["_measurement"] == "verify" and r["_field"] == "method")
  |> aggregateWindow(every: 10s, fn:count, createEmpty: false)

```
