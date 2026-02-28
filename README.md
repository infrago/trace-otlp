# trace-otlp

bamgoo otlp trace driver (HTTP endpoint).

```toml
[trace.otlp]
driver = "otlp"
[trace.otlp.setting]
endpoint = "http://127.0.0.1:4318/v1/traces"
timeout = "5s"
service = "my-service"
headers = { "Authorization" = "Bearer token" }
```
