# yeller-golang

[Yeller](http://yellerapp.com) notifier for Golang.

## Usage

In your program initialization:

```go
func main() {
        // Defaults to "production" environment
        yeller.Start("YOUR_API_KEY")

        // Or set an environment name for your application
        yeller.StartEnv("YOUR_API_KEY", "staging")
}
```

When handling errors that you'd like to log to Yeller:

```go
file, err := os.Open("filename.ext")
if err != nil {
        yeller.Notify(err)
        yeller.NotifyInfo(err, ...)
        log.Fatal(err)
}
```

if you're inside an http handler, yeller
can log other information as well:

```go
file, err := os.Open("filename.ext")
if err != nil {
        yeller.NotifyHTTP(err, request)
        yeller.NotifyHTTPInfo(err, ...)
        log.Fatal(err)
}
```
