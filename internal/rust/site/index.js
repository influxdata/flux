import { parse } from "@influxdata/flux-parser";

const INITIAL_FLUX_SCRIPT = `
from(bucket:"telegraf/autogen")
  |> limit(limit:100, offset:10)
  |> filter(fn: (r) => r.foo and r.bar or r.buz)
`;

function parseFromTextarea() {
  const source = document.querySelector("#flux-src").value.trim();
  const ast = parse(source);

  console.log("Parsed AST:", ast);

  document.querySelector("#ast").innerHTML = JSON.stringify(ast, null, 2);
}

document.querySelector("#parse").addEventListener("click", parseFromTextarea);
document.querySelector("#flux-src").value = INITIAL_FLUX_SCRIPT.trim();
parseFromTextarea();
