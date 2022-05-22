# bakelite
pure Go SQLite file exporter

This library writes SQLite files from scratch. You hand it the data you want in the tables, and you get back your .SQLite file.  
No depencencies. No C. No SQL.


# use case

For [work](https://www.engagespark.com) we have customers who have campaigns with millions of SMSs and IVRs. The details can be downloaded in an [.XLSX](https://github.com/alicebob/streamxlsx) file, but dealing with 500Mb files in Excel is no fun. Bakelite gives a light way to generate an .sqlite file.


# example

```
    db := bakelite.New()
    db.Add("planets", []string{"name", "moons"}, [][]any{
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

Not ready for production. Right now it can write basic files and sqlite is happy with those files, but there are some major basic cases missing.

main todos:
  - row encoding (we only deal with ints and strings)
  - deal with a larger number of tables (1st page mess)
  - hasn't seen a profiler. Got to make it work correctly first
  - this keeps everything in memory multiple times (and doesn't care about allocations)
  - you need to hand in all the data at once. Which won't work if you want to encode a few million rows. I can see a `db.AddCh("planets", []string{"name", "moons"}, <-chan []any)` work.
  - store as file. SQLite files can never be pure streaming, since we need to write some stuff at the beginning of the file once we have all the data, but a temp file would work, and then there are no memory restrictions anymore.

What this library won't do:
  - add indexes. (but every row gets an internal "row id", which we could expose as "integer primary key" data type, since sqlite uses the rowid for those)
  - concurrency. This library generates the file, which you can then save and use concurrently, as any normal sqlite database file, but while the file is being generated it can't be used.



## links

https://sqlite.org/fileformat2.html  
https://litecli.com/  