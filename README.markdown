# yeller-golang

[Yeller](http://yellerapp.com) notifier for Golang.

## Usage

In your program initialization:

```go
func main() {
        yeller.Start("YOUR_API_KEY")
}
```

When handing errors that you'd like to log to Yeller:

```go
file, err := os.Open("filename.ext")
if err != nil {
        yeller.Notify(err)
        log.Fatal(err)
}
```
