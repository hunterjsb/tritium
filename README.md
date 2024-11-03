# Tritium

Tritium is a secure, RAM-only storage system with zero dependencies. See [tritium-wails](https://github.com/we-be/tritium-wails) for the desktop client.

## Installation

```bash
go get github.com/we-be/tritium
```

## Quick Start

```go
import (
    "github.com/we-be/tritium/pkg/client"
)

func main() {
    c, err := client.NewClient(client.Config{
        Address: "localhost:8080",
    })
    if err != nil {
        panic(err)
    }

    // Use the client...
}
```

## Security

Tritium emphasizes security through:
- RAM-only storage
- Zero dependencies
- Secure memory management
- Optional client-side encryption
