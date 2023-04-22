document.getElementById("add-url-form").addEventListener("submit", async function (event) {
    event.preventDefault();

    const urlInput = document.getElementById("url");
    const resultDiv = document.getElementById("result");

    const response = await fetch("/add-url", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ url: urlInput.value })
    });

    if (response.ok) {
        const data = await response.json();
        resultDiv.innerHTML = `Full URL: <a href="${data.full_url}" target="_blank">${data.full_url}</a>`;
    } else {
        resultDiv.textContent = "Error adding URL";
    }
});
