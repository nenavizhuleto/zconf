# ZConf

## Subscribe for changes to ZooKeeper Nodes

```go
c.OnNodeChanged(func(path string) {
    fmt.Println("path", path, "changed")
})

c.OnNodeDataChanged(func(path string, data []byte) {
    fmt.Println("path", path, "data", string(data))
})

c.WatchPath("/")
```

see [examples](/examples) for more.
