package extensions

// TemplatePlugin is the interface to implement with custom template plugins
type TemplatePlugin interface {

	// Function is the function to be called from the template
	Function() interface{}
}
