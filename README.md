# gpuctl

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Platform: Windows | Linux](https://img.shields.io/badge/platform-Windows%20%7C%20Linux-blue)](https://github.com/zhhc99/gpuctl)
[![Go Version](https://img.shields.io/github/go-mod/go-version/zhhc99/gpuctl)](https://github.com/zhhc99/gpuctl)
[![Latest Release](https://img.shields.io/github/v/release/zhhc99/gpuctl)](https://github.com/zhhc99/gpuctl/releases)

![gpuctl image](./docs/images/demo.png)

<!-- https://ray.so/#code=JCBncHVjdGwgbGlzdApJRCAgICAgTkFNRSAgICAgICAgICAgICAgICAgICAgICAgICBVVUlEICAgICAgICBCQUNLRU5ECm46MCAgICBOVklESUEgR2VGb3JjZSBSVFggNDA2MCAgICAgIEdQVS0wMTIzLi4gIE5WTUwKaTowICAgIEludGVsIEFyYyBQcm8gQjUwICAgICAgICAgICAgR1BVLTQ1NjcuLiAgTGV2ZWxaZXJvCgokIGdwdWN0bCBnZXQgLS1kZXZpY2UgbjowLGk6MApJRCAgICAgTkFNRSAgICAgICAgICAgICBURU1QICAgRkFOICAgICAgICAgICBQT1dFUiAgICBVVElMICAgICAgICAgQ0xPQ0sgICAgICAgICBNRU1PUlkKbjowICAgIFJUWCA0MDYwICAgICAgICAgNDjCsEMgICAzNyUvMTc5MnJwbSAgIDUyVyAgICAgIEc6MCUgTTo0OCUgICBHOjIxMCBNOjQwNSAgIDIuMC84LjIgR0IKaTowICAgIEFyYyBQcm8gQjUwICAgICAgNDLCsEMgICAzMCUvMTAyN3JwbSAgIDI4VyAgICAgIEc6MCUgTToxMiUgICBHOjI0MCBNOjIxMCAgIDAuNS8xNi4wIEdCCgokIGdwdWN0bCB0dW5lIHNldCBwbD0xMDUgY29ncHU9MTgwIC1kIG46MApEZXZpY2UgbjowIChOVklESUEgR2VGb3JjZSBSVFggNDA2MCk6CiAgW-KclF0gY2xvY2tfb2Zmc2V0X2dwdT0xODAKICBb4pyUXSBwb3dlcl9saW1pdD0xMDUK&language=shell&padding=16&background=false&darkMode=true -->

**GPU 状态管理和监控工具.**

## 🛠 功能

- [x] 查看功耗, 温度, 风扇, 频率, 内存状态
- [x] 调整功耗, 频率
- [ ] 调整风扇
- [x] 配置文件, 支持登录时自动应用

支持的 GPU:

- [x] NVIDIA (NVML)
- [ ] Intel (Level Zero)
- [ ] 不会有 AMD, 直到有跨平台的 API (to AMD: Intel 比你们起步晚, 但 API 仍然跨平台)

## 🚀 快速安装

- Linux

  ```bash
  curl -sSL https://raw.githubusercontent.com/zhhc99/gpuctl/main/install.sh | sudo bash
  ```

- Windows

  ```powershell
  powershell -ExecutionPolicy ByPass -Command "iwr -useb https://raw.githubusercontent.com/zhhc99/gpuctl/main/install.ps1 | iex"
  ```

- 也可以用 `go install`:

  ```bash
  # 💡 gpuctl service install 会自动向 /usr/local/bin/ 拷贝自身, 以避免服务权限问题.
  go install github.com/zhhc99/gpuctl@latest
  ```

**卸载:**

- Linux

  ```bash
  curl -sSL https://raw.githubusercontent.com/zhhc99/gpuctl/main/uninstall.sh | sudo bash
  ```

- Windows
  - 删除文件.
  - 执行 `taskschd.msc` (任务计划程序), 删除所有前缀为 "gpuctl@" 的任务 (如果有).

## 📖 基础用法

**查看所有 GPU 状态:**

```bash
gpuctl get

# on UNIX: watch -n 1 gpuctl get
```

**设置 100w 功耗墙, 只应用到编号为 0 的 NVIDIA GPU:**

```bash
gpuctl tune set pl=100 -d n:0
```

**核心超频 +210MHz, 降压使得核心频率不超过 2520MHz:**

```bash
gpuctl tune set cogpu=210 clgpu=2520
```

> ⚠️ 一般**不认为**超频损伤硬件, 但**激进**参数可能导致**花屏**或**冻结**.

**编辑配置文件:**

```bash
gpuctl config edit
```

更多用法见 `gpuctl --help`.

> Windows 下不太容易在同一个终端提权. 考虑给 Windows Terminal 添加一个新配置文件, 并启用 "以管理员身份运行此配置文件", 之后可以在这个终端执行 `gpuctl`.

## 🔨 编译源代码

这是一个纯 go 项目, 编译非常简单.

```bash
go build -o gpuctl .
```

如果希望带有版本号:

```bash
go build -ldflags "-X 'gpuctl/cmd.Version=v1.0.0'"
```
