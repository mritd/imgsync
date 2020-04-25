# imgsync

A docker image sync tool.

|Registry|Address|Status|
|--------|-------|------|
|Flannel|[quay.io/coreos/flannel](https://quay.io/coreos/flannel)|[![](https://github.com/mritd/imgsync/workflows/Sync%20Flannel/badge.svg)](https://github.com/mritd/imgsync/actions)|
|kubeadm|[k8s.gcr.io](https://k8s.gcr.io)|[![](https://github.com/mritd/imgsync/workflows/Sync%20Kubeadm/badge.svg)](https://github.com/mritd/imgsync/actions)|
|Helm|[gcr.io/kubernetes-helm](https://gcr.io/kubernetes-helm)|[![](https://github.com/mritd/imgsync/workflows/Sync%20Helm/badge.svg)](https://github.com/mritd/imgsync/actions)|
|Istio|[gcr.io/istio-release](https://gcr.io/istio-release)|[![](https://github.com/mritd/imgsync/workflows/Sync%20Istio/badge.svg)](https://github.com/mritd/imgsync/actions)|
|Linkerd|[gcr.io/linkerd-io](https://gcr.io/linkerd-io)|[![](https://github.com/mritd/imgsync/workflows/Sync%20Linkerd/badge.svg)](https://github.com/mritd/imgsync/actions)|
|Spinnaker|[gcr.io/spinnaker-marketplace](https://gcr.io/spinnaker-marketplace)|[![](https://github.com/mritd/imgsync/workflows/Sync%20Spinnaker/badge.svg)](https://github.com/mritd/imgsync/actions)|
|Distroless|[gcr.io/distroless](https://gcr.io/distroless)|[![](https://github.com/mritd/imgsync/workflows/Sync%20Distroless/badge.svg)](https://github.com/mritd/imgsync/actions)|

# 特性

- **不依赖 Docker 运行**
- **基于 Manifests 同步**
- **支持 [Fat Manifests](https://medium.com/@arunrajeevan/handling-multi-platform-deployment-using-manifest-file-in-docker-317736a2a039) 镜像同步**
- **Manifests 文件本地 Cache，按需同步**
- **同步期间不占用本地磁盘空间(直接通过标准库转发镜像)**
- **可控的并发同步(优雅关闭/可调节并发数量)**
- **按批次同步，支持同步指定区间段镜像**
- **支持多仓库同步(后续仓库增加请提交 issue)**

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


## 其他说明

**、工具目前仅支持同步到 Docker Hub，且以后没有同步到其他仓库打算。同步 Docker Hub
时默认会同步到 `--user` 指定的用户下；镜像名称会被转换，原镜像地址内 `/` 全部转换为 `_`。
其他更细节使用请自行通过 `--help` 查看以及参考本项目 [Github Action](https://github.com/mritd/imgsync/tree/master/.github/workflows) 配置文件。**
