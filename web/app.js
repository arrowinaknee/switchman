const port = ":3315"

const baseUrl = new URL(location.protocol + location.hostname + port)

const code = document.getElementById("code")
const statusText = document.getElementById("status")

async function fetchConfig() {
	let url = new URL("/config", baseUrl)

	let response = await fetch(url)
	let source = await response.text()

	code.value = source
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

async function pressVevify() {
	let result = await verifyConfig(code.value)
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

fetchConfig()
