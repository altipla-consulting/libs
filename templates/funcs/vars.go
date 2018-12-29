package funcs

type Var struct {
	value interface{}
}

func NewVar(value interface{}) *Var {
	return &Var{value}
}

func SetVar(v *Var, value interface{}) string {
	v.value = value

	return ""
}

func GetVar(v *Var) interface{} {
	return v.value
}
