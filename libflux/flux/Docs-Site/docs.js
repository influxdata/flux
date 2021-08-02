import * as wasm from '/Users/barweiner/GolandProjects/flux/libflux/flux/pkg"';


export function get_json_documentation() {
    let identifier = document.getElementById("text").value
    let documentation = wasm.get_json_documentation(identifier);
    document.getElementById("body").innerHTML= "Documentation: " + documentation;

}