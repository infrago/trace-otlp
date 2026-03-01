# trace-otlp

`trace-otlp` 是 `trace` 模块的 `otlp` 驱动。

## 安装

```bash
go get github.com/infrago/trace@latest
go get github.com/infrago/trace-otlp@latest
```

## 接入

```go
import (
    _ "github.com/infrago/trace"
    _ "github.com/infrago/trace-otlp"
    "github.com/infrago/infra"
)

func main() {
    infra.Run()
}
```

## 配置示例

```toml
[trace]
driver = "otlp"
```

## 公开 API（摘自源码）

- `func (d *otlpDriver) Connect(inst *trace.Instance) (trace.Connection, error)`
- `func (c *otlpConnection) Open() error`
- `func (c *otlpConnection) Close() error { return nil }`
- `func (c *otlpConnection) Write(spans ...trace.Span) error`

## 排错

- driver 未生效：确认模块段 `driver` 值与驱动名一致
- 连接失败：检查 endpoint/host/port/鉴权配置
