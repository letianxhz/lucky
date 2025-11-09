package main
import _ "lucky/server/pkg/xdb/storage/mysql"
import "lucky/server/pkg/xdb"
import _ "lucky/server/pkg/xdb/example/pb"
import "fmt"
func main() {
    src := xdb.GetSourceByNS("player")
    if src == nil {
        fmt.Println("Source not found")
        return
    }
    fmt.Printf("Fields count: %d\n", len(src.Fields))
    for i, field := range src.Fields {
        fmt.Printf("  Field %d: Code=%d, Name=%s, Key=%v\n", i, field.Code, field.Name, field.Key)
    }
    names := src.FieldNames(false)
    fmt.Printf("Field names: %v\n", names)
}
