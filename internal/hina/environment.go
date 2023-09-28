package hina

type Environment struct {
	SymbolTable map[string]Term
}

func NewEnvironment() Environment {
	var env Environment
	env.SymbolTable = make(map[string]Term)
	return env
}

func (env Environment) Get(identifier string) (Term, bool) {
	node, hasNode := env.SymbolTable[identifier]
	if !hasNode {
		return nil, false
	}
	return node, true
}

func (env Environment) Set(identifier string, value Term) {
	env.SymbolTable[identifier] = value
}

func (env Environment) Copy(target Environment) {
	for key, value := range target.SymbolTable {
		if _, exists := env.Get(key); exists {
			continue
		}
		env.Set(key, value)
	}
}
