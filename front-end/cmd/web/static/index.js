(function () {
    const brokerBtn = document.querySelector("#brokerBtn");
    const output = document.querySelector("#output");
    const payload = document.querySelector("#payload");
    const received = document.querySelector("#received");

    brokerBtn.addEventListener("click", () => {
        const body = { method: "POST" };

        fetch("http://localhost:8080", body)
            .then((response) => response.json())
            .then((data) => {
                payload.innerHTML = "empty post response";
                received.innerHTML = JSON.stringify(data, undefined, 4);
                if (data.error) {
                    console.error(data.message);
                    console.log(data.message);
                } else {
                    output.innerHTML += `<br><strong>Response from broker service: </strong> ${data.message}`;
                }
            })
            .catch((error) => {
                console.error(error);
                output.innerHTML += `<br><br>Error: ${error}`;
            })
    });
})();