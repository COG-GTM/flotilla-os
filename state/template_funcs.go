package state

import (
	"text/template"

	"github.com/Masterminds/sprig"
)

// templateFuncAllowlist lists the Sprig helpers available to template
// command strings. Templates are rendered inside the API process, so anything
// that reads the process environment (env, expandenv), touches the network
// (getHostByName), generates keys, or can allocate unbounded memory/CPU
// (repeat, until, bcrypt, genCA, ...) is intentionally excluded.
var templateFuncAllowlist = []string{
	// Strings.
	"trim", "trimAll", "trimSuffix", "trimPrefix", "upper", "lower", "title",
	"untitle", "substr", "nospace", "trunc", "abbrev", "abbrevboth", "initials",
	"wrap", "wrapWith", "contains", "hasPrefix", "hasSuffix", "quote", "squote",
	"cat", "indent", "nindent", "replace", "plural", "snakecase", "camelcase",
	"kebabcase", "swapcase", "split", "splitList", "splitn", "join", "sortAlpha",
	"toString", "toStrings",
	// Regular expressions.
	"regexMatch", "regexFind", "regexFindAll", "regexReplaceAll",
	"regexReplaceAllLiteral", "regexSplit",
	// Math and conversion.
	"add", "add1", "sub", "div", "mod", "mul", "max", "min", "biggest", "ceil",
	"floor", "round", "atoi", "int", "int64", "float64",
	// Defaults and flow.
	"default", "empty", "coalesce", "ternary", "compact", "fail",
	// Encoding.
	"b64enc", "b64dec", "b32enc", "b32dec", "toJson", "toPrettyJson",
	"toRawJson", "sha1sum", "sha256sum", "adler32sum",
	// Lists and dicts.
	"list", "tuple", "first", "rest", "last", "initial", "append", "push",
	"prepend", "concat", "reverse", "uniq", "without", "has", "slice", "dict",
	"get", "set", "unset", "hasKey", "pluck", "keys", "pick", "omit", "values",
	"merge", "mergeOverwrite", "deepCopy",
	// Types and comparison.
	"typeOf", "typeIs", "typeIsLike", "kindOf", "kindIs", "deepEqual",
	// Paths (pure string manipulation).
	"base", "dir", "clean", "ext", "isAbs",
}

// TemplateFuncMap returns the restricted function map used to parse and render
// Template command strings.
func TemplateFuncMap() template.FuncMap {
	all := sprig.TxtFuncMap()
	funcs := make(template.FuncMap, len(templateFuncAllowlist))
	for _, name := range templateFuncAllowlist {
		if fn, ok := all[name]; ok {
			funcs[name] = fn
		}
	}
	return funcs
}

// ParseCommandTemplate parses a Template command string with the restricted
// function map; it fails for any function outside the allowlist.
func ParseCommandTemplate(commandTemplate string) (*template.Template, error) {
	return template.New("command").Funcs(TemplateFuncMap()).Parse(commandTemplate)
}
