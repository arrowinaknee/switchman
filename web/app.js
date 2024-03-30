const port = ":3315"

const baseUrl = new URL(location.protocol + location.hostname + port)

const code = document.getElementById("code")

async function fetchConfig() {
	let testUrl = new URL("/config", baseUrl)
	
	let response = await fetch(testUrl)
	let source = await response.text()

	code.innerHTML = source
}

fetchConfig()
