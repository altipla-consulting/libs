package funcs

import (
	"fmt"
	"io/ioutil"
	"sync"
)

var (
	includeLock  = new(sync.RWMutex)
	includeCache = map[string]string{}
)

func Include(path string) (string, error) {
	includeLock.RLock()
	cache := includeCache[path]
	includeLock.RUnlock()

	if cache != "" && !Development() {
		return cache, nil
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("templates: cannot include file: %v", path)
	}

	includeLock.Lock()
	defer includeLock.Unlock()
	includeCache[path] = string(content)

	return string(content), nil
}
