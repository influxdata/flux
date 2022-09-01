use serde_json::Value;
#[allow(unreachable_code)]
pub fn process_response_flux(response: &str) -> Result<(), anyhow::Error> {
    if let Ok(a) = serde_json::from_str::<Value>(response) {
        //flux result

        println!(
            "\n{}",
            serde_json::to_string(&a["result"]["Result"])?.replace('\"', "")
        );
        return Ok(());
    } else {
        //error case
        println!("{}", response);
        return Ok(());
    }
    Ok(())
}
