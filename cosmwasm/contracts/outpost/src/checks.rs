use crate::ContractError;

pub fn validate_input_amount(
    input_amount: impl Into<u128>,
    sent_amount: impl Into<u128>,
) -> Result<(), ContractError> {
    let input_amount = input_amount.into();
    let sent_amount = sent_amount.into();

    if input_amount > sent_amount {
        return Err(ContractError::SwapAmountTooHigh {
            received: input_amount,
            max: sent_amount,
        });
    };
    Ok(())
}
