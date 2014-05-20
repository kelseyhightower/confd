package template

// StoreClient is used to swap out the backend store.
type StoreClient interface {
	GetValues(keys []string) (map[string]interface{}, error)
}
