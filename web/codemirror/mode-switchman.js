CodeMirror.defineSimpleMode("switchman", {
	start: [
		{regex: /#.*/, token: "comment"},
		{regex: /(\w+)(\s*{)/, token: ["keyword", null]},
		{regex: /((?:"(?:[^"]|\\")*"|'(?:[^']|\\')*'|[^\s:{}"']+))(\s*:\s*)/, token: ["variable", "operator"]},
		{regex: /(?:"(?:[^"]|\\")*"|'(?:[^']|\\')*'|[^\s:{}"']+)/, token: "string"},
		{regex: /[^\s{}:]+/, token: "error"},
	],
	meta: {
		lineComment: "#",
	}
})
