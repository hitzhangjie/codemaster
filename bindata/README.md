# bindata 轻松实现对静态资源的二进制打包

go似乎有计划要支持对静态资源文件的打包，但是目前最新go发行版仍然不支持。

如果想实现类似功能，就只能借助一些第三方的库，比如go-rice、go-bindata等等，这些三方库用起来也没那么简单方便。

这里写了个简单的demo，后续有类似需求可以考虑这么实现，后面go官方支持了再考虑切换实现方式，没必要用那些过度设计的三方库了。

**1. 演示下如何使用bindata来将文件或者整个目录转换成go文件存储**

```go
# 构建bindata工具
go build -v

# 将整个static目录压缩后存储到一个gobin/static.go文件
./bindata -input static/ -output gobin/static.go
```

**2. 演示下如何在代码中使用上述生成的go文件来还原回以前的文件内容**

- 代码中其他地方直接通过 `gobin.StaticGo` 来引用上述内容
- 可以通过 `compress.UnTar(dst, bytes.NewBuffer(gobin.StaticGo))`解压到本地文件系统，然后就能当做本地文件来使用

利用这种转换，开发者可以很方便地将一些依赖本地配置文件、静态资源文件的项目做下优化。
可以轻松实现go get安装，您可以选择在程序首次执行时将打包好的静态资源文件释放到本地文件系统中去。
