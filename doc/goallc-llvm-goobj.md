# GoALLC：LLVM IR 到 GoObj 的开发链路

本文记录当前已合入的、用于开发 Go SSA 到 LLVM lowering 的最小端到端
链路。它让一个简单 Go package 经由 LLVM IR、`llc` 和 GoObj 进入正常的
Go package archive，最终仍由社区 Go linker 链接。

这是一条受限的开发路径，不是社区 Go 后端的替代品。完整的项目背景、构建
基线和更早的试验记录见 [../GOALLC.md](../GOALLC.md)。

## 数据流

```text
go build -toolexec=llvmtoolexec
  -> llvmtoolexec 拦截选中的 compile 调用
  -> compile -enablellvm -llvmironly
       -> __.PKGDEF + <archive>.ll
  -> llc -filetype=obj <archive>.ll
       -> LLVM GoObj
  -> cmd/internal/archive 将 GoObj 追加为 _go_.o
  -> cmd/link 读取 __.PKGDEF 和 _go_.o
```

`-llvmironly` 必须和 `-enablellvm` 一起使用。它使 `compile` 在写出 LLVM
IR 和 Go archive 的 `__.PKGDEF` 后结束，不再生成原生 linker object。因而
对被替换的 package 而言，archive 中唯一的 `_go_.o` 必须是 `llc` 生成的
GoObj；`__.PKGDEF` 仍完全由 Go compiler 生成并位于 archive 的第一个成员。

## 类型描述符和只读数据

`-llvmironly` 不会跳过 compiler 的 `dumpdata` 准备阶段。`reflectdata` 仍按
原生 compiler 的方式在 `base.Ctxt.Data` 中生成最终的 `obj.LSym` 布局；随后
LLVM lowering 从带有 `TypeInfo` 的 runtime type descriptor roots 出发，收集
仅由这些 roots 的 relocation 可达的数据闭包，并将其 lower 为 LLVM constant
globals。descriptor 的主体由同一份 `rttype` runtime ABI layout 构造为 packed
named LLVM structs（例如 `%go.runtime.Type`、`%go.runtime.StructType`）；未建
schema 的辅助数据仍以 bytes/relocations 表示。这个边界刻意位于 finalized LSym，
而不是重写 reflectdata 布局逻辑。

每个被 lower 的 LSym 保留：

- descriptor 已知字段的 runtime ABI 类型、字段 offset、原始值和
  `.rodata` / `.data` / `.bss` 段属性；
- `R_ADDR`、`R_ADDROFF`、`R_METHODOFF` 及其 weak 变体的 addend 和目标；
- `DUPOK`、`LOCAL`、typelink、Go type、itab、`UsedInIface`、linkname 等 GoObj
  symbol flags；
- runtime ABIInternal 函数引用（例如 type equality closure 中的
  `runtime.memequal64`）。

LLVM IR 使用 global attachment `!goobj.symbol.flags` 和 `!goobj.relocs` 交接
这些 Go-only 属性。GoObj writer 消费该 metadata，在默认 relocation 推导前恢复
精确的 Go relocation type，特别是 weak offset relocation。零宽度的 linker
保活边 `R_KEEP` 不伪装为地址常量，而用 `!goobj.keep` 记录其 target symbol，
由 writer 合成 GoObj relocation。最终链路仍只有
`__.PKGDEF + _go_.o` 两个有意义的 archive members；不会生成或合并 native data
object。

静态 interface conversion 还会把带有 `ItabInfo` 的 roots 纳入同一数据闭包。
itab 使用 `%go.runtime.ITab` 和按实际方法数扩展的 packed LLVM struct 表达；
固定的 `Fun[0]` 后面按 LSym 的最终大小追加 `uintptr` 数组，因此单方法和多方法
itab 都保持 runtime ABI 的连续方法表布局。方法入口仍保留 weak `R_ADDR`
relocation。interface 方法调用从 itab 槽加载入口，转成 LLVM function pointer，
并以原 SSA `AuxCall` 的 ABIInternal signature 发出 indirect call。

Go linker 的 dead-method elimination 还依赖函数上的零宽度
`R_USEIFACE`、`R_USEIFACEMETHOD` 和 `R_USENAMEDMETHOD`。它们不对应 LLVM
机器指令或地址常量，因此 compiler 用 `!goobj.marker_relocs` 保存精确的
relocation type、addend 和 target name；GoObj writer 将其恢复到源函数的
relocation 列表。普通 LLVM call graph 不能替代这项 linker 契约。

non-empty interface 到另一 non-empty interface 的转换沿用 compiler 生成的
`internal/abi.TypeAssert` cache 和 `runtime.typeAssert` fallback。LLVM lowering
保留 cache 的 sequentially-consistent pointer load、pointer/uintptr probe 比较、
nil 分支和 ABIInternal runtime call；cache miss 与随后命中的 fast path 均由
runtime 测试覆盖。empty interface 到 concrete type 的 comma-ok assertion
直接比较动态 type word 并按 SSA 结果形状取回 data；empty interface 到
non-empty interface 的 comma-ok assertion 复用同一 TypeAssert cache/fallback，
并覆盖成功、类型不匹配和 nil 输入。panic-form assertion 沿用相同 fast path，
失败边分别调用 `runtime.panicdottypeE`、`runtime.panicdottypeI` 或
`runtime.panicnildottype`。

interface type switch 对 concrete cases 使用动态 type hash/type pointer 比较；
interface case 使用 compiler 生成的 `internal/abi.InterfaceSwitch` descriptor、
atomic cache probe 和 `runtime.interfaceSwitch` fallback。descriptor 作为带
`AuxGotype` 的可写 data root 进入同一 LSym-to-LLVM data closure。

这类可写 descriptor 还要求 GoObj symbol 的精确大小、对齐和 `AuxGotype`。
LLVM GoObj writer 从 global layout 写入 symbol size/alignment，排除 section
padding，并通过 `!goobj.gotype` 恢复 compiler LSym 的 Go type auxiliary。
该 auxiliary 的 target 既可以是本包定义，也可以是 external non-package
reference；后者用于 builtin 或其他 package 拥有的 type symbol。
否则 linker 无法为 `.data` 生成正确的 GC bitmap。

当前实现只 lower type-rooted data closure，尚未把所有 `dumpdata` 产物泛化为
LLVM data lowering。type descriptor 中未被当前 schema 覆盖的尾部数据，以及其他
data roots，保守地维持 bytes/relocation fallback；新增 schema 或 root 前必须先明确
其 relocation、GC 和 linker 契约，并为 writer 增加相应的 GoObj regression。

## LLVM IR 契约

LLVM IR 本身是 frontend 到 `llc` 的唯一配置载体。`compile` 设置 GoObj
target triple，并添加一个名为 `!goobj.config` 的 named metadata：

```llvm
target triple = "aarch64-apple-darwin-goobj"
!goobj.config = !{!0}
!0 = !{!"goallc.goobj", !"darwin", !"arm64", !"go1.27rc2",
       !"GOARM64", !"v8.0", !"build-id", !"main", !"1", !"0", !1}
!1 = !{!"regabiwrappers", !"regabiargs"}
```

`!0` 的字段顺序固定，每一项均为独立 metadata operand：

| Index | 字段 |
| --- | --- |
| 0 | schema 标识，固定为 `goallc.goobj` |
| 1–2 | `GOOS`、`GOARCH` |
| 3 | Go compiler 版本 |
| 4–5 | 架构设置的 key/value，例如 `GOARM64` / `v8.0` |
| 6 | build ID |
| 7 | package path |
| 8–9 | `main`、`shared`，值为 `0` 或 `1` |
| 10 | experiment metadata node；其中每个 operand 是一个 experiment 名称 |

experiments 不使用逗号分隔字符串，其他字段也不拼接成 Go textual object
header。`llc` 在建立 code-generation pipeline 前读取并严格校验这些字段，
然后把它们传给 GoObj writer。没有 `!goobj.config` 的普通 GoObj IR 仍可使用
既有 `-goobj-*` command-line 选项作为兼容回退。

对应 LLVM 测试为
`llvm/test/CodeGen/AArch64/goobj-ir-config.ll`。LLVM 侧实现位于
`llvm/tools/llc/llc.cpp`、`llvm/lib/CodeGen/CodeGenTargetMachineImpl.cpp`
和 `llvm/lib/MC/GoObjObjectWriter.cpp`。

## toolexec wrapper

`cmd/llvmtoolexec` 只处理 `compile`，其余 Go tool 调用原样透传。它：

1. 为选中的 compile 调用增加 `-enablellvm -llvmironly`；
2. 调用 `llc -filetype=obj`，不传递或重建 GoObj header 配置；
3. 使用 `cmd/internal/archive` 打开 compiler archive，并将对象以 `_go_.o`
   追加进去；
4. 默认删除 IR 和临时 object，`-keep-ir` 可保留 IR 供检查。

wrapper 不解析 `__.PKGDEF`、不反向识别 Go textual header，也不自行写 ar
header。使用 Go 自己的 archive writer 可保证 `__.PKGDEF` 保持第一个 member，
避免 BSD `ar` 插入 `__.SYMDEF` 后破坏 `cmd/link` 的读取约定。

可用 `-package` 或环境变量 `GOALLC_LLVM_PACKAGE` 限定 import path；未设置时
所有 `compile` 调用都会走这条实验链路。`-llc` 或 `GOALLC_LLC` 必须指向带有
GoObj metadata 支持的 GoALLC LLVM 构建。

## 最小使用方式

假定 `$GOROOT` 是本仓库构建出的 Go toolchain，`$LLC` 是对应 LLVM 分支构建
的 `llc`：

```sh
cd "$GOROOT/src/cmd"
go build -o /private/tmp/llvmtoolexec ./llvmtoolexec

cd /path/to/simple-main-package
GOALLC_LLVM_PACKAGE=main \
  "$GOROOT/bin/go" build \
  -toolexec="/private/tmp/llvmtoolexec -llc=$LLC -keep-ir" \
  -o app .
./app
```

在 shell 中也可设置 `GOALLC_LLC="$LLC"`，这样 `-toolexec` 只需指定 wrapper
路径。`-keep-ir` 留下的 IR 路径为 compile 输出 archive 路径加 `.ll` 后缀。

开发时至少运行：

```sh
cd "$GOROOT/src/cmd"
go test ./llvmtoolexec

cd /path/to/llvm-project
llvm/cmake-build-debug/bin/llvm-lit -sv \
  llvm/test/CodeGen/AArch64/goobj-ir-config.ll
```

此外应以一个简单 `main` package 执行真实的 `go build -toolexec`，并运行产物。
这同时验证 compiler archive、IR metadata、`llc`、GoObj archive member 和
`cmd/link` 的接口。

## 回归测试机制

GoALLC 不维护一套与 Go 仓库重复的测试源码。LLVM 测试由
`cmd/internal/testdir` 发现并复用 `$GOROOT/test` 中已有的测试：

- codegen 候选是 `test/codegen` 下 recipe 为 `// asmcheck` 的文件；
- runtime 候选是 testdir 原本扫描目录中 recipe 为 `// run` 的文件；
- `test/llvm_tests.json` 分别维护 codegen 和 runtime 的白名单与黑名单；
- 白名单是当前必须通过的用例；黑名单支持 glob，用于记录尚未覆盖的范围；
  精确白名单优先于宽泛黑名单；
- runner 会拒绝拼错的白名单、无匹配项的黑名单，以及未被任一名单分类的
  候选，并报告当前白名单文件数和候选文件总数。

### LLVM IR codegen 检查

codegen 白名单文件直接在原 Go 源码中使用 LLVM `FileCheck` 指令，例如：

```go
// LLVM-DAG: define goabiinternal { i64, i8 } @codegen.div_ndivis6_int64
// LLVM-DAG: sdiv i64
// LLVM-DAG: srem i64
```

runner 使用 `compile -enablellvm -llvmironly` 生成 `.ll`，然后执行：

```sh
FileCheck --check-prefix=LLVM test/codegen/example.go < package.a.ll
```

因此 codegen 测试只检查 LLVM IR，不调用 `llc`、不链接也不运行。可使用
`LLVM:`、`LLVM-LABEL:`、`LLVM-DAG:`、`LLVM-NOT:` 等标准 FileCheck
指令。原有 asmcheck parser 会忽略这些独立注释行，社区 Go 的机器码检查仍
按原方式运行。每个 codegen 白名单文件至少要包含一条 `// LLVM...`
指令。

`FileCheck` 的选择顺序为 `GOALLC_FILECHECK`、
`$GOALLC_LLVM_DIR/bin/FileCheck`、`$GOROOT/llvm/bin/FileCheck`。

### LLVM runtime 检查

runtime 白名单不增加新的 recipe，也不复制测试源码。runner 仍解析原文件的
`// run` 参数、build constraint、超时和期望输出，但构建步骤改为通过
`cmd/llvmtoolexec` 仅替换 `main` package：

```text
原有 // run 用例
  -> go build -toolexec="llvmtoolexec -package=main ..."
  -> compile -enablellvm -llvmironly
  -> llc 生成 _go_.o
  -> cmd/link
  -> 运行并沿用原 testdir 输出检查
```

`llc` 的选择顺序为 `GOALLC_LLC`、`$GOALLC_LLVM_DIR/bin/llc`、
`$GOROOT/llvm/bin/llc`。

### 扩大覆盖范围

新增覆盖时：

1. 先确认整个现有测试文件都能由当前 LLVM lowering 处理；
2. codegen 文件在源码中增加针对 LLVM IR 的 FileCheck 指令；
3. 在 `test/llvm_tests.json` 中增加精确白名单项和简短能力说明；
4. runtime 文件只需加入白名单，不修改其原有 `// run` recipe；
5. 运行 LLVM 定向测试，并同时运行对应的原生 asmcheck/run 测试。

定向运行全部 LLVM 白名单：

```sh
go test cmd/internal/testdir -run='^Test$/^LLVM$' -v
```

只运行一个 LLVM codegen 或 runtime 子测试时，可继续在 `-run` 中追加
对应的 slash-separated subtest 名称。黑名单用于明确尚未覆盖的范围，不应
把实际失败的用例加入白名单后再按 expected failure 处理。

## 当前范围与后续工作

- 当前 target triple 仅配置了 `darwin/arm64` 和 `linux/amd64`。
- LLVM SSA lowering 仍不完整；复杂 SSA op、GC pointer liveness、defer/panic、
  closure、完整 ABI/DWARF 等尚未达到通用正确性。
- 当前 interface 范围包括 compiler 能静态生成 itab 的 concrete-to-interface
  conversion、对应的 ABIInternal 间接方法调用，以及 non-empty
  interface-to-interface conversion；empty-interface 到 concrete/non-empty
  interface 的 comma-ok/panic-form assertion，以及包含 concrete/interface cases
  的 type switch 也已覆盖。
- 每个经 LLVM 替换的 package 都必须由 `llc` 生成完整的 `_go_.o`，不能与
  原生 compiler object 混合。
- `!goobj.config` 是这条开发链路的稳定交接点。新增 header 配置时应新增独立
  metadata 字段和相应 `llc` 验证，不要恢复 wrapper 中的 header 解析逻辑。
- 类型 descriptor data 已覆盖，但这不等同于通用 static-data lowering；目前仅
  支持 type-rooted readonly/data closure 中已验证的 relocation 类型。

相关合入记录：Go [#3](https://github.com/goallc/go/pull/3)，LLVM
[#3](https://github.com/goallc/llvm-project/pull/3)。
