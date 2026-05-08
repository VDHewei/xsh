# Network Proxy Configuration

## Local Proxy

When accessing external networks times out, or when accessing foreign websites fails, use the local proxy:

```powershell
$Env:http_proxy="http://127.0.0.1:63459";$Env:https_proxy="http://127.0.0.1:63459"
```

For bash:

```bash
export http_proxy="http://127.0.0.1:63459"
export https_proxy="http://127.0.0.1:63459"
```

## Usage

- GitHub downloads (e.g., onnxruntime-genai library)
- Go module downloads from foreign registries
- Any external HTTP/HTTPS requests that timeout without proxy
