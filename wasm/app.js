const statusEl = document.getElementById("status");
const outputEl = document.getElementById("output");
const btnGreet = document.getElementById("btn-greet");
const btnAdd = document.getElementById("btn-add");

function log(message) {
  outputEl.textContent += message + "\n";
}

async function loadGoWasm() {
  if (!window.Go) {
    throw new Error("wasm_exec.js not loaded");
  }

  const go = new Go();
  const result = await WebAssembly.instantiateStreaming(
    fetch("main.wasm"),
    go.importObject
  );

  go.run(result.instance);

  statusEl.textContent = "WASM ready";
  btnGreet.disabled = false;
  btnAdd.disabled = false;

  btnGreet.addEventListener("click", () => {
    const message = greet("Browser");
    log(message);
  });

  btnAdd.addEventListener("click", () => {
    const sum = add(1, 2);
    log(`add(1, 2) = ${sum}`);
  });
}

loadGoWasm().catch((err) => {
  statusEl.textContent = "Failed to load WASM";
  log(String(err));
  console.error(err);
});
