const js = import("@influxdata/parser");

function parse_flux() {
    src =  document.querySelector("#flux_src").value;
    js.then(js => {
        // TODO the current parser can't handle trailing whitespace so we just trim it.
        ast = js.js_parse(src.trim());
        console.log("Parsed AST", ast);
        document.querySelector("#ast").innerHTML = JSON.stringify(ast, null, 2);
    });
}

document.querySelector('#parse').addEventListener('click', parse_flux)
