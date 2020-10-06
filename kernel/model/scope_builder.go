package model

func BuildScope() *ScopeBuilder {
	return &ScopeBuilder{}
}

type ScopeBuilder struct {
	Scope
}

func (b *ScopeBuilder) Build() Scope {
	return b.Scope
}

func (b *ScopeBuilder) Var(name ...string) *VarBuilder {
	if len(name) == 0 {
		panic("variables must have at least one component in the path")
	}
	newVar := b.Variables.NewVariable(name...)
	return &VarBuilder{
		ScopeBuilder: b,
		currentVar:   newVar,
	}
}

type VarBuilder struct {
	*ScopeBuilder
	currentVar *Variable
}

func (b *VarBuilder) Default(val interface{}) *VarBuilder {
	b.currentVar.Default = val
	return b
}

func (b *VarBuilder) Required() *VarBuilder {
	b.currentVar.Required = true
	return b
}

func (b *VarBuilder) Sensitive() *VarBuilder {
	b.currentVar.Sensitive = true
	return b
}
