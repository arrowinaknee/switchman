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

let edit_user = ""

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

async function fetchUsers() {
	let url = new URL("/users", baseUrl)

	let response = await fetch(url)
	let users = await response.json()
	console.log(users)

	let table = document.getElementById("user-list")
	let tbody = table.tBodies[0]
	for (let row of Array.from(tbody.children)) {
		if (!row.classList.contains("static")) {
			row.remove()
		}
	}

	if (users.length == 0) {
		let row = document.createElement("tr")
		row.innerHTML = `<td colspan="3">No users found</td>`
		tbody.appendChild(row)
		return
	}
	for (let user of users) {
		let row = document.createElement("tr")
		row.innerHTML = `
			<td>${user.login}</td>
			<td>Enabled</td>
			<td class="phover row-button">Manage ></td>`
		row.classList = "hover-bg clickable"
		row.onclick = () => showUser(user.login)
		tbody.appendChild(row)
	}
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

async function userCreate() {
	let login = document.getElementById("newuser-login").value
	let password = document.getElementById("newuser-password").value
	let passwordConfirm = document.getElementById("newuser-password-confirm").value

	if (password != passwordConfirm) {
		return // TODO: show error
	}

	let user = {
		"login": login,
		"password": password,
	}

	let url = new URL("/users", baseUrl)

	let response = await fetch(url, {
		method: "post",
		body: JSON.stringify(user)
	})
	if (response.status == 200) {
		showPanelUsers()
	}
}

function userLoginRefresh() {
	let login = document.getElementById("user-login").value
	console.log(login, edit_user)
	let button = document.getElementById("user-login-update")
	// console.log(button.disabled)
	// button.disabled = (login == "" || login == edit_user) ? true : false
	// console.log(button.disabled)
	if (login == "" || login == edit_user) {
		button.setAttribute("disabled", true)
	} else {
		button.removeAttribute("disabled")
	}
}

function userPasswordRefresh() {
	let password = document.getElementById("user-password").value
	let passwordConfirm = document.getElementById("user-password-confirm").value
	console.log(password, passwordConfirm)
	let button = document.getElementById("user-password-update")
	button.disabled = (password == "" || password != passwordConfirm) ? true : false
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
	fetchUsers()
}

function showCreateUser() {
	showPanel("panel-newuser")
	document.getElementById("newuser-login").value = ""
	document.getElementById("newuser-password").value = ""
	document.getElementById("newuser-password-confirm").value = ""
}

function showUser(login) {
	edit_user = login
	document.getElementById("user-login").value = login
	showPanel("panel-user")
	userLoginRefresh()
}

showPanelCode()
