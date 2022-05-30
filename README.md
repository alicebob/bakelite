# bakelite

Pure Go SQLite file exporter

This library writes SQLite files from scratch. You hand it the data you want in
the tables, and you get back your .SQLite file.  

No dependencies. No C. No SQL.


# use case

At [work](https://www.engagespark.com), we have customers who have campaigns
with millions of SMSs and IVRs. The details can be downloaded as an
[.XLSX](https://github.com/alicebob/streamxlsx) file, but dealing with 500Mb
files in Excel is no fun. Bakelite gives a light way to generate a .sqlite
file.


# example

```
    db := bakelite.New()
    // Table with all data from a slice
    err := db.AddSlice("planets", []string{"name", "moons"}, [][]any{
        {"Mercury", 0},
        {"Venus", 0},
        {"Earth", 1},
        {"Mars", 2},
        {"Jupiter", 80},
        {"Saturn", 83},
        {"Uranus", 27},
        {"Neptune", 4},
    })

    // Table with all data from a channel
    stars := make(chan []any, 10)
    stars <- []any{"Alpha Centauri", "4"}
    stars <- []any{"Barnard's Star", "6"}
    stars <- []any{"Luhman 16", "6"}
    stars <- []any{"WISE 0855−0714", "7"}
    stars <- []any{"Wolf 359", "7"}
    err := db.AddChan("stars", []string{"name", "lightyears"}, stars)

    b := &bytes.Buffer{}
    err := db.WriteTo(b)
```


# status

Not used in production yet. It can write files and SQLite is happy with those
files.

Main todos:
  - support more Go datatypes, such as int32 and bool.
  - hasn't seen a profiler.

What this library won't do:
  - add indexes. (but every row gets an internal "row id", which we could
    expose as "integer primary key" data type, since sqlite uses the rowid for
    those)
  - concurrency. This library generates the file, which you can then save and
    use concurrently, as any normal sqlite database file. But while the file is
    being generated, it can't be used.
  - updates. This is a write-once affair.


## links

https://sqlite.org/fileformat2.html  
https://litecli.com/  
