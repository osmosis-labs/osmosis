use crate::bindings::GoString;
use std::ffi::CString;

/// conversion from &CString to GoString
impl From<&CString> for GoString {
    fn from(c_str: &CString) -> Self {
        let ptr_c_str = c_str.as_ptr();

        GoString {
            p: ptr_c_str,
            n: c_str.as_bytes().len() as isize,
        }
    }
}

/// This is needed to be implemented as macro since
/// conversion from &CString to GoString requires
/// CString to not get dropped before referecing its pointer
#[macro_export]
macro_rules! redefine_as_go_string {
    ($($ident:ident),*) => {
        $(
            let $ident = &std::ffi::CString::new($ident).unwrap();
            let $ident: $crate::bindings::GoString = $ident.into();
        )*
    };
}
