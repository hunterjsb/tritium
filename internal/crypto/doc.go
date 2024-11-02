package crypto // import "github.com/hunterjsb/tritium/pkg/crypto"

/*
Package crypto provides secure encryption primitives for Tritium.

Basic usage:

    key := make([]byte, 32)
    rand.Read(key)

    block, err := crypto.NewBlock(key, data)
    if err != nil {
        log.Fatal(err)
    }

    encrypted := block.Bytes()
*/
