package main

import (
    "bytes"
    "fmt"
    "reflect"
)

func main() {
    s1 := []byte("Hello Bytes")

    // 分配一个buffer
    buf := bytes.NewBuffer(s1)
    // 如果是指针就会报错
    buf_t := reflect.TypeOf(*buf)
    fmt.Printf("buf type[%s] name[%s] kind[%s], num_fields[%d] num_method[%d]\n", 
        buf_t.String(), buf_t.Name(), buf_t.Kind(), buf_t.NumField(), buf_t.NumMethod())

    // 得到所有成员
    for i := 0; i < buf_t.NumField(); i++ {
        field := buf_t.Field(i)
        fmt.Printf("field:%d name[%s] type[%s]\n", i, field.Name, field.Type.Name())
    }

    fmt.Printf("buf len[%d] cap[%d] content[%s]\n", buf.Len(), buf.Cap(), buf)

    // 是否可以直接修改里面的内容
    ref_buf := buf.Bytes()
    ref_buf[0] = 'c'
    ref_buf[1] = 'c'
    fmt.Printf("after modify Bytes buf len[%d] cap[%d] buf[%s]\n", buf.Len(), buf.Cap(), buf)

    b, _ := buf.ReadByte()
    fmt.Printf("after ReadByte buf len[%d] cap[%d] b=[%c] buf[%s]\n", buf.Len(), buf.Cap(), b, buf)

    read_buf := make([]byte, 8)
    ret, _ := buf.Read(read_buf)
    fmt.Printf("buf len[%d] cap[%d] ret[%d] read_buf[%s] buf[%s]\n", buf.Len(), buf.Cap(), ret, read_buf, buf)

    // 读取完毕之后在写入，肯定会增加cap，其内部会判断，如果已经空了，其内部会自己做truncate，这样就会从头写，而不会扩展
    ret, _ = buf.Write(read_buf)
    fmt.Printf("buf len[%d] cap[%d] ret[%d] buf[%s]\n", buf.Len(), buf.Cap(), ret, buf)
}