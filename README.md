# Tritium

Tritium is a secure, RAM-only storage system with zero dependencies. See [tritium-wails](https://github.com/we-be/tritium-wails) for the desktop client.

![Screenshot from 2024-11-06 21-44-02](https://github.com/user-attachments/assets/2a00124f-78f7-4721-bb8b-d70bf6733446)


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

## Persistence and Scaling
Read the proposal [here](https://gist.github.com/hunterjsb/572f8e3b66dde9551e3fa3652f6b40b7)
