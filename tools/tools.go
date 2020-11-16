// +build tools

// the build flag is needed to avoid side-effects of init() functions in tool dependencies.

// Go modules support developer tools (commands) as dependencies.
// For example, your project might require a tool to help with code generation, or to lint/vet your code for correctness.
// Adding developer tool dependencies ensures that all developers use the same version of each tool.

package tools

import (
	// exists solely for its side effects, in this case the declaration of the dependency.
	_ "github.com/kisielk/errcheck"
)
