use crate::runner::error::{DecodeError, RunnerError};
use cosmrs::proto::cosmos::base::abci::v1beta1::{GasInfo, TxMsgData};
use cosmrs::proto::tendermint::abci::ResponseDeliverTx;
use cosmwasm_std::{Attribute, Event};
use prost::Message;
use std::ffi::CString;
use std::str::Utf8Error;

pub type RunnerResult<T> = Result<T, RunnerError>;
pub type RunnerExecuteResult<R> = Result<ExecuteResponse<R>, RunnerError>;

#[derive(Debug, Clone, PartialEq)]
pub struct ExecuteResponse<R>
where
    R: prost::Message + Default,
{
    pub data: R,
    pub raw_data: Vec<u8>,
    pub events: Vec<Event>,
    pub gas_info: GasInfo,
}

impl<R> TryFrom<ResponseDeliverTx> for ExecuteResponse<R>
where
    R: prost::Message + Default,
{
    type Error = RunnerError;

    fn try_from(res: ResponseDeliverTx) -> Result<Self, Self::Error> {
        let tx_msg_data =
            TxMsgData::decode(res.data.as_slice()).map_err(DecodeError::ProtoDecodeError)?;

        let msg_data = &tx_msg_data
            .data
            // since this tx contains exactly 1 msg
            // when getting none of them, that means error
            .get(0)
            .ok_or(RunnerError::ExecuteError { msg: res.log })?;

        let data = R::decode(msg_data.data.as_slice()).map_err(DecodeError::ProtoDecodeError)?;

        let events = res
            .events
            .into_iter()
            .map(|e| -> Result<Event, DecodeError> {
                Ok(Event::new(e.r#type.to_string()).add_attributes(
                    e.attributes
                        .into_iter()
                        .map(|a| -> Result<Attribute, Utf8Error> {
                            Ok(Attribute {
                                key: std::str::from_utf8(a.key.as_slice())?.to_string(),
                                value: std::str::from_utf8(a.value.as_slice())?.to_string(),
                            })
                        })
                        .collect::<Result<Vec<Attribute>, Utf8Error>>()?,
                ))
            })
            .collect::<Result<Vec<Event>, DecodeError>>()?;

        Ok(ExecuteResponse {
            data,
            raw_data: res.data,
            events,
            gas_info: GasInfo {
                gas_wanted: res.gas_wanted as u64,
                gas_used: res.gas_used as u64,
            },
        })
    }
}

/// `RawResult` facilitates type conversions between Go and Rust,
///
/// Since Go struct could not be exposed via cgo due to limitations on
/// its unstable behavior of its memory layout.
/// So, apart from passing primitive types, we need to:
///
///   Go { T -> bytes(T) -> base64 -> *c_char }
///                      ↓
///   Rust { *c_char -> base64 -> bytes(T') -> T' }
///
/// Where T and T' are corresponding data structures, regardless of their encoding
/// in their respective language plus error information.
///
/// Resulted bytes are tagged by prepending 4 bytes to byte array
/// before base64 encoded. The prepended byte represents
///   0 -> Ok
///   1 -> QueryError
///   2 -> ExecuteError
///
/// The rest are undefined and remaining spaces are reserved for future use.
#[derive(Debug)]
pub struct RawResult(Result<Vec<u8>, RunnerError>);

impl RawResult {
    /// Convert ptr to AppResult. Check the first byte tag before decoding the rest of the bytes into expected type
    pub(crate) fn from_ptr(ptr: *mut std::os::raw::c_char) -> Option<Self> {
        if ptr.is_null() {
            return None;
        }

        let c_string = unsafe { CString::from_raw(ptr) };
        let base64_bytes = c_string.to_bytes();
        let bytes = base64::decode(base64_bytes).unwrap();
        let code = bytes[0];
        let content = &bytes[1..];

        if code == 0 {
            Some(Self(Ok(content.to_vec())))
        } else {
            let content_string = CString::new(content)
                .unwrap()
                .to_str()
                .expect("Go code must encode valid UTF-8 string")
                .to_string();

            let error = match code {
                1 => RunnerError::QueryError {
                    msg: content_string,
                },
                2 => RunnerError::ExecuteError {
                    msg: content_string,
                },
                _ => panic!("undefined code: {}", code),
            };
            Some(Self(Err(error)))
        }
    }

    /// Convert ptr to AppResult. Use this function only when it is sure that the
    /// pointer is not a null pointer.
    ///
    /// # Safety
    /// There is a potential null pointer here, need to be extra careful before
    /// calling this function
    pub(crate) unsafe fn from_non_null_ptr(ptr: *mut std::os::raw::c_char) -> Self {
        Self::from_ptr(ptr).expect("Must ensure that the pointer is not null")
    }

    pub(crate) fn into_result(self) -> Result<Vec<u8>, RunnerError> {
        self.0
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::account::Account;
    use crate::runner::app::OsmosisTestApp;
    use crate::runner::error::RunnerError::{ExecuteError, QueryError};
    use crate::runner::Runner;
    use osmosis_std::types::osmosis::gamm::poolmodels::balancer::v1beta1::{
        MsgCreateBalancerPool, MsgCreateBalancerPoolResponse,
    };
    use osmosis_std::types::osmosis::gamm::v1beta1::{
        PoolParams, QueryPoolRequest, QueryPoolResponse,
    };

    #[derive(::prost::Message)]
    struct AdhocRandomQueryRequest {
        #[prost(uint64, tag = "1")]
        id: u64,
    }

    #[derive(::prost::Message)]
    struct AdhocRandomQueryResponse {
        #[prost(string, tag = "1")]
        msg: String,
    }

    #[test]
    fn test_query_error_no_route() {
        let app = OsmosisTestApp::default();
        let res = app.query::<AdhocRandomQueryRequest, AdhocRandomQueryResponse>(
            "/osmosis.random.v1beta1.Query/AdhocRandom",
            &AdhocRandomQueryRequest { id: 1 },
        );

        let err = res.unwrap_err();
        assert_eq!(
            err,
            QueryError {
                msg: "No route found for `/osmosis.random.v1beta1.Query/AdhocRandom`".to_string()
            }
        );
    }

    #[test]
    fn test_query_error_failed_query() {
        let app = OsmosisTestApp::default();
        let res = app.query::<QueryPoolRequest, QueryPoolResponse>(
            "/osmosis.gamm.v1beta1.Query/Pool",
            &QueryPoolRequest { pool_id: 1 },
        );

        let err = res.unwrap_err();
        assert_eq!(
            err,
            QueryError {
                msg: "rpc error: code = Internal desc = pool with ID 1 does not exist".to_string()
            }
        );
    }

    #[test]
    fn test_execute_error() {
        let app = OsmosisTestApp::default();
        let signer = app.init_account(&[]).unwrap();
        let res: RunnerExecuteResult<MsgCreateBalancerPoolResponse> = app.execute(
            MsgCreateBalancerPool {
                sender: signer.address(),
                pool_params: Some(PoolParams {
                    swap_fee: "10000000000000000".to_string(),
                    exit_fee: "10000000000000000".to_string(),
                    smooth_weight_change_params: None,
                }),
                pool_assets: vec![],
                future_pool_governor: "".to_string(),
            },
            MsgCreateBalancerPool::TYPE_URL,
            &signer,
        );

        let err = res.unwrap_err();
        assert_eq!(
            err,
            ExecuteError {
                msg: String::from("pool should have at least 2 assets, as they must be swapping between at least two assets")
            }
        )
    }

    #[test]
    fn test_raw_result_ptr_with_0_bytes_in_content_should_not_error() {
        let base64_string = base64::encode(vec![vec![0u8], vec![0u8]].concat());
        let res = RawResult::from_ptr(
            CString::new(base64_string.as_bytes().to_vec())
                .unwrap()
                .into_raw(),
        )
        .unwrap()
        .into_result()
        .unwrap();

        assert_eq!(res, vec![0u8]);
    }
}
