#[macro_export]
macro_rules! fn_execute {
    (pub $name:ident: $req:ty[$type_url:expr] => $res:ty) => {
        pub fn $name(
            &self,
            msg: $req,
            signer: &$crate::SigningAccount,
        ) -> $crate::RunnerExecuteResult<$res> {
            self.runner.execute(msg, $type_url, signer)
        }
    };
    (pub $name:ident: $req:ty => $res:ty) => {
        pub fn $name(
            &self,
            msg: $req,
            signer: &$crate::SigningAccount,
        ) -> $crate::RunnerExecuteResult<$res> {
            self.runner.execute(msg, <$req>::TYPE_URL, signer)
        }
    };
    ($name:ident: $req:ty[$type_url:expr] => $res:ty) => {
        pub fn $name(
            &self,
            msg: $req,
            signer: &$crate::SigningAccount,
        ) -> $crate::RunnerExecuteResult<$res> {
            self.runner.execute(msg, $type_url, signer)
        }
    };
    ($name:ident: $req:ty => $res:ty) => {
        pub fn $name(
            &self,
            msg: $req,
            signer: &$crate::SigningAccount,
        ) -> $crate::RunnerExecuteResult<$res> {
            self.runner.execute(msg, <$req>::TYPE_URL, signer)
        }
    };
}

#[macro_export]
macro_rules! fn_query {
    (pub $name:ident [$path:expr]: $req:ty => $res:ty) => {
        pub fn $name(&self, msg: &$req) -> $crate::RunnerResult<$res> {
            self.runner.query::<$req, $res>($path, msg)
        }
    };
    ($name:ident [$path:expr]: $req:ty => $res:ty) => {
        fn $name(&self, msg: &$req) -> $crate::RunnerResult<$res> {
            self.runner.query::<$req, $res>($path, msg)
        }
    };
}
