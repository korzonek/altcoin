package types

type Hasher interface {
	Hash() string
}

type Serializer interface {
	JSON() string
}
