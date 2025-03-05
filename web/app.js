const port = ":3315"

const baseUrl = new URL(location.protocol + location.hostname + port)

const codeAttach = document.getElementById("code-hook")
const code = CodeMirror(codeAttach, {
	lineNumbers: true,
	indentWithTabs: true,
	mode: 'switchman',
	theme: 'material-darker',
	scrollbarStyle: "null",
	autoIndent: false,
})
const statusText = document.getElementById("status")

function StatusOK(text) {
	statusText.innerHTML = text
	if (statusText.classList.contains("error"))
		statusText.classList.remove("error")
}
function StatusError(text) {
	statusText.innerHTML = text
	if (!statusText.classList.contains("error"))
		statusText.classList.add("error")
}

async function fetchConfig() {
	let url = new URL("/config", baseUrl)

	let response = await fetch(url)
	let source = await response.text()

	code.setValue(source)
}

async function verifyConfig(code) {
	let url = new URL("/verify", baseUrl)

	let response = await fetch(url, {
		method: "post",
		body: code
	})
	let status = await response.text()
	return status
}

async function pressVerify() {
	let result = await verifyConfig(code.getValue())
	if (result == "") {
		statusText.innerHTML = "Config is valid"
		if (statusText.classList.contains("error"))
			statusText.classList.remove("error")
	} else {
		statusText.innerHTML = result
		if (!statusText.classList.contains("error"))
			statusText.classList.add("error")
	}
}

async function updateConfig(code) {
	let url = new URL("/config", baseUrl)

	let response = await fetch(url, {
		method: "post",
		body: code
	})
	let status = await response.text()
	return status
}

async function pressApply() {
	let result = await updateConfig(code.getValue())
	if (result == "") {
		StatusOK("Config applied successfully")
	} else {
		StatusError(result)
	}
}

// async function pressLogin() {
// 	let login = document.getElementById("login").value
// 	let password = document.getElementById("password").value

// 	let response = await fetch("/login", {
// 		method: "post",
// 		body: JSON.stringify({
// 			"login": login,
// 			"password": password
// 		})
// 	})
// 	let status = await response.text()
// 	console.log(status)
// }

function showPanel(id) {
	let panels = Array.from(document.getElementById("panel-mount").children)
	let active = panels.filter(p => p.classList.contains("active"))
	console.log(active)
	for (let p of active) {
		p.classList.remove("active")
	}
	let panel = document.getElementById(id)
	panel.classList.add("active")
}

function showPanelCode() {
	showPanel("panel-code")
	fetchConfig()
}

function showPanelUsers() {
	showPanel("panel-users")
}

showPanel("panel-code")
