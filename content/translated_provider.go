package content

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type TranslatedProvider map[string]Translated

// SetValue changes a single value in a single provider. It helps to avoid a nil
// panic when the provider is not assigned yet.
func (content TranslatedProvider) SetValue(provider, lang, value string) {
	if content[provider] == nil {
		content[provider] = map[string]string{}
	}
	content[provider][lang] = value
}

// Provider returns a single provider from the data. If the provider is not
// assigned yet it will return an empty translation instead of the nil of a direct access.
func (content TranslatedProvider) Provider(provider string) Translated {
	if content[provider] == nil {
		return Translated{}
	}

	return content[provider]
}

// Chain returns the value of every language of the first provider of the global chain list
// that has content.
func (content TranslatedProvider) Chain() Translated {
	return content.CustomChain(globalProviderChain)
}

// ChainProvider returns the first provider of every language of the global chain list
// that has content.
func (content TranslatedProvider) ChainProvider() Translated {
	return content.CustomChainProvider(globalProviderChain)
}

// CustomChain returns the value of every language of the first provider of the chain list
// that has content. If no chain is provided it returns a random provider.
//
// Any provider not in the list won't count the in the chain at all, it will be ignored.
func (content TranslatedProvider) CustomChain(chain []string) Translated {
	if len(chain) == 0 {
		for _, v := range content {
			return v
		}
	}

	inverseChain := make([]string, 0, len(chain))
	for i := len(chain) - 1; i >= 0; i-- {
		inverseChain = append(inverseChain, chain[i])
	}

	result := make(Translated)
	for _, p := range inverseChain {
		for lang, value := range content[p] {
			if value != "" {
				result[lang] = value
			}
		}
	}

	return result
}

// CustomChainProvider returns the first provider of every language of the chain list
// that has content. If no chain is provided it returns a random provider.
//
// Any provider not in the list won't count the in the chain at all, it will be ignored.
func (content TranslatedProvider) CustomChainProvider(chain []string) Translated {
	if len(chain) == 0 {
		for p, v := range content {
			reply := make(Translated)
			for lang := range v {
				reply[lang] = p
			}
			return reply
		}
	}

	inverseChain := make([]string, 0, len(chain))
	for i := len(chain) - 1; i >= 0; i-- {
		inverseChain = append(inverseChain, chain[i])
	}

	result := make(Translated)
	for _, p := range inverseChain {
		for lang, value := range content[p] {
			if value != "" {
				result[lang] = p
			}
		}
	}

	return result
}

func (content TranslatedProvider) Value() (driver.Value, error) {
	if content == nil {
		return "{}", nil
	}

	serialized, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("content/translated-provider: cannot serialize value: %s", err)
	}

	return serialized, nil
}

func (content *TranslatedProvider) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("content/translated-provider: cannot scan type into bytes: %T", value)
	}

	if err := json.Unmarshal(b, content); err != nil {
		return fmt.Errorf("content/translated-provider: cannot scan value: %s", err)
	}

	if *content == nil {
		*content = TranslatedProvider{}
	}

	return nil
}
