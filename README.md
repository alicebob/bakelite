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
    db.AddSlice("planets", []string{"name", "moons"}, [][]any{
        {"Mercury", 0},
        {"Venus", 0},
        {"Earth", 1},
        {"Mars", 2},
        {"Jupiter", 80},
        {"Saturn", 83},
        {"Uranus", 27},
        {"Neptune", 4},
    })

    b := &bytes.Buffer{}
    db.Write(b)
```


# status

Not ready for production. It can write files and SQLite is
happy with those files, but it's still early code.

Main todos:
  - row encoding (we only deal with ints and strings)
  - hasn't seen a profiler. Got to make it work correctly first
  - this keeps everything in memory multiple times (and doesn't care about allocations)
  - store as file. SQLite files can never be pure streaming, since we need to
    write some stuff at the beginning of the file once we have all the data,
    but a temp file would work, and then there are no memory restrictions anymore.

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
