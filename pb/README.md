旧版本的protoc-gen-go生成的*.pb.go桩代码里面有一个FileDescriptor_${id}的玩意，如果将一个协议hello直接拷贝一份hellox，并且修改package名
hello为hellox和注册的类型名hello.HelloReq为hellox.HelloReq，是不够的，还需要改这里的${id}，最好还是重新protoc生成一遍。

这里如果同时import的话，就会导致冲突，这当然是以为protoc-gen-go处理的不够健壮。

在新版本的protoc-gen-go插件中，已经修复了此问题，生成的桩代码中已经废弃了FileDescriptor_${id}注册的逻辑。