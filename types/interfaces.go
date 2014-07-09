package types

type Hasher interface {
	Hash() string
}

type Serializer interface {
	Json() string
}
