use crate::account::SigningAccount;
use crate::runner::result::{RunnerExecuteResult, RunnerResult};

pub mod app;
pub mod error;
pub mod result;

pub trait Runner<'a> {
    fn execute<M, R>(
        &self,
        msg: M,
        type_url: &str,
        signer: &SigningAccount,
    ) -> RunnerExecuteResult<R>
    where
        M: ::prost::Message,
        R: ::prost::Message + Default,
    {
        self.execute_multiple(&[(msg, type_url)], signer)
    }

    fn execute_multiple<M, R>(
        &self,
        msgs: &[(M, &str)],
        signer: &SigningAccount,
    ) -> RunnerExecuteResult<R>
    where
        M: ::prost::Message,
        R: ::prost::Message + Default;

    fn query<Q, R>(&self, path: &str, query: &Q) -> RunnerResult<R>
    where
        Q: ::prost::Message,
        R: ::prost::Message + Default;
}
