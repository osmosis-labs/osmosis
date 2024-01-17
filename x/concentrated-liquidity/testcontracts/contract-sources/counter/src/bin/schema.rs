use cosmwasm_schema::write_api;

use counter::msg::InstantiateMsg;

fn main() {
    write_api! {
        instantiate: InstantiateMsg,
    }
}
