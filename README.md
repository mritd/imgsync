![imgsync](.logo.png)

## imgsync

A docker image sync tool.

|Registry|Address|Docker Hub|Status|
|--------|-------|----------|------|
|Flannel|[quay.io/coreos/flannel](https://quay.io/coreos/flannel)|`gcrxio/quay.io_coreos_flannel`|[![Build Status](https://travis-ci.org/mritd/imgsync.svg?branch=master)](https://travis-ci.org/mritd/imgsync)|
|kubeadm|[k8s.gcr.io](https://k8s.gcr.io)|`gcrxio/k8s.gcr.io_*`|[![Build Status](https://travis-ci.org/mritd/imgsync.svg?branch=master)](https://travis-ci.org/mritd/imgsync)|
|Helm|[gcr.io/kubernetes-helm](https://gcr.io/kubernetes-helm)|`gcrxio/gcr.io_kubernetes-helm_*`|[![Build Status](https://travis-ci.org/mritd/imgsync.svg?branch=master)](https://travis-ci.org/mritd/imgsync)|
|Istio|[gcr.io/istio-release](https://gcr.io/istio-release)|`gcrxio/gcr.io_istio-release_*`|[![Build Status](https://travis-ci.org/mritd/imgsync.svg?branch=master)](https://travis-ci.org/mritd/imgsync)|
|Linkerd|[gcr.io/linkerd-io](https://gcr.io/linkerd-io)|`gcrxio/gcr.io_linkerd-io_*`|[![Build Status](https://travis-ci.org/mritd/imgsync.svg?branch=master)](https://travis-ci.org/mritd/imgsync)|
|Spinnaker|[gcr.io/spinnaker-marketplace](https://gcr.io/spinnaker-marketplace)|`gcrxio/gcr.io_spinnaker-marketplace_*`|[![Build Status](https://travis-ci.org/mritd/imgsync.svg?branch=master)](https://travis-ci.org/mritd/imgsync)|
|Distroless|[gcr.io/distroless](https://gcr.io/distroless)|`gcrxio/gcr.io_distroless_*`|[![Build Status](https://travis-ci.org/mritd/imgsync.svg?branch=master)](https://travis-ci.org/mritd/imgsync)|
|Samples|[gcr.io/google-samples](https://gcr.io/google-samples)|`gcrxio/gcr.io_google-samples_*`|[![Build Status](https://travis-ci.org/mritd/imgsync.svg?branch=master)](https://travis-ci.org/mritd/imgsync)|
|KNative|[gcr.io/knative-releases](https://gcr.io/knative-releases)|`gcrxio/gcr.io_knative-releases_*`|[![Build Status](https://travis-ci.org/mritd/imgsync.svg?branch=master)](https://travis-ci.org/mritd/imgsync)|

**如何快速拉取 kubeadm 镜像?**

```sh
for img in `kubeadm config images list`; do
    docker pull "gcrxio/$(echo $img | tr '/' '_')" && docker tag "gcrxio/$(echo $img | tr '/' '_')" $img;
done
```

## 特性

- **不依赖 Docker 运行**
- **基于 Manifests 同步**
- **支持 [Fat Manifests](https://medium.com/@arunrajeevan/handling-multi-platform-deployment-using-manifest-file-in-docker-317736a2a039) 镜像同步**
- **Manifests 文件本地 Cache，按需同步**
- **同步期间不占用本地磁盘空间(直接通过标准库转发镜像)**
- **可控的并发同步(优雅关闭/可调节并发数量)**
- **按批次同步，支持同步指定区间段镜像**
- **支持多仓库同步(后续仓库增加请提交 issue)**
- **支持生成同步报告，同步报告推送 [Telegram](https://t.me/imgsync)**

## 安装

工具采用 go 编写，安装可直接从 release 页下载对应平台二进制文件，并增加可执行权限运行既可

```bash
wget https://github.com/mritd/imgsync/releases/download/v2.0.0/imgsync_linux_amd64
chmod +x imgsync_linux_amd64
./imgsync_linux_amd64 --help
```

## 编译

如果预编译不包含您的平台架构，可以自行编译本工具，本工具编译依赖如下

- make
- git
- Go 1.14.2+

编译时直接运行 `make bin` 命令既可，如需交叉编译请安装 [gox](https://github.com/mitchellh/gox) 工具并执行 `make` 命令既可

## 使用

```bash
Docker image sync tool.

Usage:
  imgsync [flags]
  imgsync [command]

Available Commands:
  flannel     Sync flannel images
  gcr         Sync gcr images
  help        Help about any command
  sync        Sync single image

Flags:
      --debug     debug mode
  -h, --help      help for imgsync
  -v, --version   version for imgsync

Use "imgsync [command] --help" for more information about a command..
```

### sync

`sync` 子命令用于同步单个镜像，一般用于测试目的进行同步并查看相关日志

### gcr

`gcr` 子命令用户同步 **gcr.io** 相关镜像，如果使用 `--kubeadm` 选项则同步 **k8s.gcr.io** 镜像

### flannel

`flannel` 子命令用于同步 **quay.io** 的 flannel 镜像

## 推荐配置

由于工具会开启并发同步，且不经过 Docker，不进行本地缓存，所以本工具推荐的最低运行配置如下:

- 4 核心 8G 内存 vps
- 至少 100/M 对等的带宽接口
- Ubuntu 18.04+ 系统环境
- 磁盘至少保留 2G 可用空间(manifests 本地缓存需要用到一定空间)

**本工具默认 20 并发进行同步处理，且每次同步针对每个镜像 tag 至少发出一次 manifest 请求；
这意味着当前(在本文档编写时)每次全部仓库同步至少发出 100083 个 manifests 请求以及其他试图
获取镜像名称列表、tag 列表的请求，在高并发下这需要服务器有足够的 CPU 和带宽能力；内存占用方面
目前还可以接受，主要内存消耗在启动时加载 manifests 配置文件并反序列化到内存 map，这期间大约
需要花费最高 10s 的时间(434M json 文件)。**

## 镜像名称

工具默认会转换原镜像名称，转换规则为将原镜像名称内的 `/` 全部替换为 `_`，例如(假设 Docker Hub 用户名为 `gcrxio`):

**`gcr.io/istio-release/pilot:latest` ==> `gcrxio/gcr.io_istio-release_pilot:latest`**

## 国内 Docker Hub Mirror

- Aliyun: `[系统分配前缀].mirror.aliyuncs.com`
- Tencent: `https://mirror.ccs.tencentyun.com`
- 163: `http://hub-mirror.c.163.com`
- Azure: `dockerhub.azk8s.cn`

## 其他说明

**工具目前仅支持同步到 Docker Hub，且以后没有同步到其他仓库打算。同步 Docker Hub
时默认会同步到 `--user` 指定的用户下；本工具默认已经将支持的仓库同步到 Docker Hub [gcrxio](https://hub.docker.com/u/gcrxio) 用户下；
其他更细节使用请自行通过 `--help` 查看以及参考本项目 [Travis CI](https://github.com/mritd/imgsync/tree/master/.travis.yml) 配置文件。
项目目录下的 Github Action 配置已经停用(性能不行)，没删除是因为调了好久删了可惜，以后备用吧。**
