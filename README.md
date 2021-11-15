# jenga
jenga是一个带索引可顺序添加(并压缩)数据的数据写入/读取工具，支持数据压缩的算法有：
* gzip
* zlib

## 1 安装
```
go get github.com/xfali/jenga/jenga
```

## 2 使用
### 2.1 压缩
jenga add

将指定文件/目录添加到jenga文件中

参数
* -j 指定生成的jenga文件路径
* -s 指定原文件/目录路径，如为目录则压缩目录中所有文件
* -k 指定关联查找/获取文件的key
* -g 指定使用的压缩算法为gzip
* -z 指定使用的压缩算法为zlib

示例：
```
jenga add -j all.ja.gz -g -s data1
```
### 2.2 查询索引
jenga list

列出所有文件的索引，通过索引可以获得该文件

参数
* -j 指定查询的jenga文件路径

示例：
```
jenga list -j all.ja.gz
```

### 2.3 获得文件
jenga get

从jenga文件中提取通过key提取指定文件，或者提取所有文件到指定目录

参数
* -j 指定jenga文件路径
* -k 指定提取文件的key(可以通过jenga list查询)
* -f 指定提取文件的目的路径（可以是文件或者目录）

示例：
```
jenga get -j all.ja.gz -k test -f test
```

## 3 项目集成

### 3.1 安装依赖
```
go get github.com/xfali/jenga
```

### 3.2 压缩文件
```
// 指定压缩算法为gzip(仅写入时需要指定)
blks = jenga.NewJenga("./test.je.gz", jenga.V2Gzip())
err := blks.Open(jenga.OpFlagCreate | jenga.OpFlagWriteOnly)
if err != nil {
    t.Fatal(err)
}
defer blks.Close()
// reader为要写入数据的io.Reader
err = blks.Write(key, -1, reader)
if err != nil {
    t.Fatal(err)
}
```

### 3.3 查询索引
```
blks = jenga.NewJenga("./test.je.gz")
err := blks.Open(jenga.OpFlagReadOnly)
if err != nil {
    t.Fatal(err)
}
defer blks.Close()
l := blks.KeyList()
for _, v := range l {
    t.Log(v)
}
```

### 3.4 根据索引提取文件
```
blks = jenga.NewJenga("./test.je.gz")
err := blks.Open(jenga.OpFlagReadOnly)
if err != nil {
    t.Fatal(err)
}
defer blks.Close()
// writer为要写入key关联数据的io.Writer
_, err = blks.Read(key, writer)
```