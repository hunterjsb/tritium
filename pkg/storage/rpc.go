package storage

type SetArgs struct {
	Key   string
	Value []byte
	TTL   *int // optional TTL in seconds
}

type SetReply struct {
	Error string
}

type GetArgs struct {
	Key string
}

type GetReply struct {
	Value []byte
	Error string
}
