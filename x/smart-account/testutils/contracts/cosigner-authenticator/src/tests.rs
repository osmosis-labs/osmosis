use sylvia::multitest::App;

use crate::contract::multitest_utils::CodeId;

#[test]
fn instantiate() {
    let app = App::default();
    let code_id = CodeId::store_code(&app);

    let owner = "owner";

    let contract = code_id.instantiate(42).call(owner).unwrap();

    let count = contract.count().unwrap().count;
    assert_eq!(count, 42);
}
