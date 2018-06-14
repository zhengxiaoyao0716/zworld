# ZWorld
## 用一个种子生成一个世界，做得到吗？

***
## Usage
[Preview](http://zworld.temp.zheng0716.com/index.html)
> Note: 抵达区块边界后请通过控制台 `debug.command.shiftChunk(ChunkID)` 命令进行区块传送（因为时间原因我没完成自动加载下一区块）<br>
> Note: 预览地址仅为单节点，部署于阿里云，单用户体验用。多用户应当部署新的系统拷贝，并加入路由以形成集群。

Download the binary: [release](https://github.com/zhengxiaoyao0716/zworld/releases)
Download the resource: [browser](https://github.com/zhengxiaoyao0716/zworld/tree/master/browser)
``` bash
./zworld --help
```

## Deploy
``` bash
./zworld service --help
```

## Build
Use `build.sh` to build both `win64/win32/linux64` versions.

## Develop
``` bash
go run zworld.go --help
```
