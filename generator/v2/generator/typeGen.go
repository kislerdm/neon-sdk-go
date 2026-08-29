package generator

type typeDefinitionInput struct {
	schema OpenAPISchema
	name   string
}

type TypesRepo struct {
	typesDefinition []string

	inputQueue []typeDefinitionInput
}

func (v *TypesRepo) AddTypeDefinitionInput(schema OpenAPISchema, name string) {
	if v.inputQueue == nil {
		v.inputQueue = make([]typeDefinitionInput, 0, 100)
	}
	v.inputQueue = append(v.inputQueue, typeDefinitionInput{
		schema: schema,
		name:   name,
	})
}

// Dequeue FIFO dequeue.
func (v *TypesRepo) dequeue() (OpenAPISchema, string) {
	var schema OpenAPISchema
	var name string

	if len(v.inputQueue) > 0 {
		schema, name = v.inputQueue[0].schema, v.inputQueue[0].name
	}

	if len(v.inputQueue) > 1 {
		v.inputQueue = v.inputQueue[1:]
	} else {
		v.inputQueue = nil
	}

	return schema, name
}

func (v *TypesRepo) EmptyInputQueue() bool {
	return len(v.inputQueue) == 0
}

func (v *TypesRepo) AddTypeDefinition(s string) {
	if v.typesDefinition == nil {
		v.typesDefinition = make([]string, 0, 100)
	}
	v.typesDefinition = append(v.typesDefinition, s)
}

func (v *TypesRepo) TypesDefinition() []string {
	return v.typesDefinition
}
