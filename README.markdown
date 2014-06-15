# yeller-golang

[Yeller](http://yellerapp.com) notifier for Golang.

# Reporting Errors to Yeller in 30 Seconds

In your program initialization:

```go
func main() {
        // Start yeller in a specific deployment environment
        yeller.StartEnv("YOUR_API_KEY", "staging")

        // Alternatively, default to "production" environment
        // yeller.Start("YOUR_API_KEY")
}
```

When handling errors that you'd like to log to Yeller:

```go
file, err := os.Open("filename.ext")
if err != nil {

        // notify the error by itself, with additional
        // information about what was going on when the error happened
        info := make(map[string]interface{})
        info["filename"] = "filename.ext"
        yeller.NotifyInfo(err, info)

        // notify the error by itself, with no additional information
        // yeller.Notify(err)
        log.Fatal(err)
}
```

Generally you should log your errors to Yeller at the highest possible level
that still has good context information about what happened.

## HTTP Error Handling
If you're inside an http handler, yeller
can log request information as well:

```go
http.HandleFunc("/foo", func(w http.ResponseWriter, req *http.Request) {
        file, err := os.Open("filename.ext")
        if err != nil {
                yeller.NotifyHTTP(err, req)
                yeller.NotifyHTTPInfo(err, ...)
                log.Fatal(err)
        }
})
```

## Handling Panic
Most golang programs never use panic intentionally, but it is essential to know
when your program panics. Yeller handles that for you, deduplicating panics so
even if you have thousands of crashes, you'll be able to distinguish between
causes.

```go
func f() {
        myFunctionThatPanics()
        defer func() {
            if r := recover(); r != nil {
                yeller.NotifyPanic(r)
            }
        }
}
```
