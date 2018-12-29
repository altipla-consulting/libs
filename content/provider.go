package content

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

var globalProviderChain = []string{}

type Provider map[string]string

// Chain returns the value of the first provider of the global chain list
// that has content. If no provider has content it will return an empty string.
func (content Provider) Chain() string {
	return content.CustomChain(globalProviderChain)
}

// ChainProvider returns the first provider of the global chain list
// that has content. If no provider has content it will return an empty string.
func (content Provider) ChainProvider() string {
	return content.CustomChainProvider(globalProviderChain)
}

// CustomChain returns the value of the first provider of the chain list
// that has content. If no chain is provided it returns a random provider. If no
// provider has content it will return an empty string.
//
// Any provider not in the list won't count the in the chain at all, it will be ignored.
func (content Provider) CustomChain(chain []string) string {
	if len(chain) == 0 {
		for _, v := range content {
			return v
		}
	}

	for _, p := range chain {
		if content[p] != "" {
			return content[p]
		}
	}

	return ""
}

// CustomChainProvider returns the first provider of the chain list
// that has content. If no chain is provided it returns a random provider. If no
// provider has content it will return an empty string.
//
// Any provider not in the list won't count the in the chain at all, it will be ignored.
func (content Provider) CustomChainProvider(chain []string) string {
	if len(chain) == 0 {
		for p := range content {
			return p
		}
	}

	for _, p := range chain {
		if content[p] != "" {
			return p
		}
	}

	return ""
}

func (content Provider) Value() (driver.Value, error) {
	if content == nil {
		return "{}", nil
	}

	serialized, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("content/provider: cannot serialize value: %s", err)
	}

	return serialized, nil
}

func (content *Provider) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("content/provider: cannot scan type into bytes: %T", value)
	}

	if err := json.Unmarshal(b, content); err != nil {
		return fmt.Errorf("content/provider: cannot scan value: %s", err)
	}

	if *content == nil {
		*content = Provider{}
	}

	return nil
}

// SetGlobalProviderChain configures the global chain of providers when calling
// a Provider.Chain() method. It is NOT thread-safe, you should call it at init()
// when starting the app and it shouldn't change again never.
func SetGlobalProviderChain(chain []string) {
	globalProviderChain = chain
}
