(function () {
  const brokerBtn = document.querySelector("#brokerBtn");
  const authBrokerBtn = document.querySelector("#authBrokerBtn");
  const logBrokerBtn = document.querySelector("#logBrokerBtn");
  const mailBrokerBtn = document.querySelector("#mailBrokerBtn");
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

  mailBrokerBtn.addEventListener("click", () => {
    const payload = {
      action: "mail",
      mail: {
        from: "me@example.com",
        to: "you@there.com",
        subject: "Test Email",
        message: "Hello world!",
      }
    };

    const headers = new Headers();
    headers.append("Content-Type", "application/json");

    const body = {
      method: "POST",
      body: JSON.stringify(payload),
      headers: headers,
    }

    fetch("http://localhost:8080/handle", body)
      .then((response) => response.json())
      .then((data) => {
        payload.innerHTML = JSON.stringify(payload, undefined, 4);
        received.innerHTML = JSON.stringify(data, undefined, 4);
        if (data.error) {
          output.innerHTML += `<br><strong>Error:</strong> ${data.message}`;
        } else {
          output.innerHTML += `<br><strong>Response from broker service: </strong> ${data.message}`;
        }
      })
      .catch((error) => {
        console.error(error);
        output.innerHTML += `<br><br>Error: ${error}`;
      })
  });

  logBrokerBtn.addEventListener("click", () => {
    const payload = {
      action: "log",
      log: {
        name: "event",
        data: "some logging stuff"
      }
    };

    const headers = new Headers();
    headers.append("Content-Type", "application/json");

    const body = {
      method: "POST",
      body: JSON.stringify(payload),
      headers: headers,
    }

    fetch("http://localhost:8080/handle", body)
      .then((response) => response.json())
      .then((data) => {
        payload.innerHTML = JSON.stringify(payload, undefined, 4);
        received.innerHTML = JSON.stringify(data, undefined, 4);
        if (data.error) {
          output.innerHTML += `<br><strong>Error:</strong> ${data.message}`;
        } else {
          output.innerHTML += `<br><strong>Response from broker service: </strong> ${data.message}`;
        }
      })
      .catch((error) => {
        console.error(error);
        output.innerHTML += `<br><br>Error: ${error}`;
      })
  });

  authBrokerBtn.addEventListener("click", () => {
    const payload = {
      action: "auth",
      auth: {
        email: "admin@example.com",
        password: "verysecret"
      }
    };

    const headers = new Headers();
    headers.append("Content-Type", "application/json");

    const body = {
      method: "POST",
      body: JSON.stringify(payload),
      headers: headers,
    }

    fetch("http://localhost:8080/handle", body)
      .then((response) => response.json())
      .then((data) => {
          payload.innerHTML = JSON.stringify(payload, undefined, 4);
          received.innerHTML = JSON.stringify(data, undefined, 4);
          if (data.error) {
              output.innerHTML += `<br><strong>Error:</strong> ${data.message}`;
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