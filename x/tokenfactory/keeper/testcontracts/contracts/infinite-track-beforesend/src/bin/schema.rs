use cosmwasm_schema::write_api;

use infinite_track_beforesend::msg::InstantiateMsg;

fn main() {
    write_api! {
        instantiate: InstantiateMsg,
    }
}
