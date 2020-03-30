package firestore

type Model interface {
	Collection() string
	Key() string
}
