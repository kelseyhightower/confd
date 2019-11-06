package template

type KeysCache struct {
	FuncMap map[string]interface{}
	Keys    []string
}

func NewKeysCache() *KeysCache {
	k := &KeysCache{}
	k.FuncMap = map[string]interface{}{
		"exists": k.Exists,
		"ls":     k.List,
		"lsdir":  k.ListDir,
		"get":    k.Get,
		"gets":   k.GetAll,
		"getv":   k.GetValue,
		"getvs":  k.GetAllValues,
	}
	return k
}

func (k *KeysCache) CacheKey(key string) {
	k.Keys = append(k.Keys, key)
}

func (k *KeysCache) Del(key string) {
	k.CacheKey(key)
}

func (k *KeysCache) Exists(key string) bool {
	k.CacheKey(key)
	return false
}

func (k *KeysCache) Get(key string) (map[string]string, error) {
	k.CacheKey(key)
	return nil, nil
}

func (k *KeysCache) GetValue(key string, v ...string) (string, error) {
	k.CacheKey(key)
	return "", nil
}

func (k *KeysCache) GetAll(pattern string) (map[string]string, error) {
	return nil, nil
}

func (k *KeysCache) GetAllValues(pattern string) ([]string, error) {
	return nil, nil
}

func (k *KeysCache) List(filePath string) []string {
	return nil
}

func (k *KeysCache) ListDir(filePath string) []string {
	return nil
}

func (k *KeysCache) Set(key string, value string) {
	k.CacheKey(key)
}

func (k *KeysCache) Purge() {
}
