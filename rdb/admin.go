package rdb

type AdminOperation interface {
	URL() string
	Body() interface{}
}

type ConfigureExpirationOperation struct {
	Disabled             bool
	DeleteFrequencyInSec *int
}

func (op *ConfigureExpirationOperation) SetFrequency(freq int) {
	op.DeleteFrequencyInSec = &freq
}

func (op *ConfigureExpirationOperation) URL() string {
	return "/admin/expiration/config"
}

func (op *ConfigureExpirationOperation) Body() interface{} {
	return op
}

func DisableExpiration() *ConfigureExpirationOperation {
	return &ConfigureExpirationOperation{
		Disabled: true,
	}
}

func EnableExpiration() *ConfigureExpirationOperation {
	return &ConfigureExpirationOperation{}
}
