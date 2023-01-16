use itertools::Itertools;
use proc_macro::TokenStream;
use proc_macro2::TokenTree;
use quote::quote;
use syn::{parse_macro_input, DeriveInput};

macro_rules! match_kv_attr {
    ($key:expr, $value_type:tt) => {
        |tt| {
            if let [TokenTree::Ident(key), TokenTree::Punct(eq), TokenTree::$value_type(value)] =
                &tt[..]
            {
                if (key == $key) && (eq.as_char() == '=') {
                    Some(quote!(#value))
                } else {
                    None
                }
            } else {
                None
            }
        }
    };
}

#[proc_macro_derive(CosmwasmExt, attributes(proto_message, proto_query))]
pub fn derive_cosmwasm_ext(input: TokenStream) -> TokenStream {
    let input = parse_macro_input!(input as DeriveInput);
    let ident = input.ident;

    let type_url = get_type_url(&input.attrs);

    // `EncodeError` always indicates that a message failed to encode because the
    // provided buffer had insufficient capacity. Message encoding is otherwise
    // infallible.

    let (query_request_conversion, cosmwasm_query) = if get_attr("proto_query", &input.attrs)
        .is_some()
    {
        let path = get_query_attrs(&input.attrs, match_kv_attr!("path", Literal));
        let res = get_query_attrs(&input.attrs, match_kv_attr!("response_type", Ident));

        let query_request_conversion = quote! {
            impl <Q: cosmwasm_std::CustomQuery> From<#ident> for cosmwasm_std::QueryRequest<Q> {
                fn from(msg: #ident) -> Self {
                    cosmwasm_std::QueryRequest::<Q>::Stargate {
                        path: #path.to_string(),
                        data: msg.into(),
                    }
                }
            }
        };

        let cosmwasm_query = quote! {
            pub fn query(self, querier: &cosmwasm_std::QuerierWrapper<impl cosmwasm_std::CustomQuery>) -> cosmwasm_std::StdResult<#res> {
                querier.query::<#res>(&self.into())
            }
        };

        (query_request_conversion, cosmwasm_query)
    } else {
        (quote!(), quote!())
    };

    (quote! {
        impl #ident {
            pub const TYPE_URL: &'static str = #type_url;
            #cosmwasm_query
        }

        #query_request_conversion

        impl From<#ident> for cosmwasm_std::Binary {
            fn from(msg: #ident) -> Self {
                let mut bytes = Vec::new();
                prost::Message::encode(&msg, &mut bytes)
                    .expect("Message encoding must be infallible");

                cosmwasm_std::Binary(bytes)
            }
        }

        impl<T> From<#ident> for cosmwasm_std::CosmosMsg<T> {
            fn from(msg: #ident) -> Self {
                cosmwasm_std::CosmosMsg::<T>::Stargate {
                    type_url: #type_url.to_string(),
                    value: msg.into(),
                }
            }
        }

        impl TryFrom<cosmwasm_std::Binary> for #ident {
            type Error = cosmwasm_std::StdError;

            fn try_from(binary: cosmwasm_std::Binary) -> Result<Self, Self::Error> {
                use ::prost::Message;
                Self::decode(&binary[..]).map_err(|e| {
                    cosmwasm_std::StdError::ParseErr {
                        target_type: stringify!(#ident).to_string(),
                        msg: format!(
                            "Unable to decode binary: \n  - base64: {}\n  - bytes array: {:?}\n\n{:?}",
                            binary,
                            binary.to_vec(),
                            e
                        ),
                    }
                })
            }
        }

        impl TryFrom<cosmwasm_std::SubMsgResult> for #ident {
            type Error = cosmwasm_std::StdError;

            fn try_from(result: cosmwasm_std::SubMsgResult) -> Result<Self, Self::Error> {
                result
                    .into_result()
                    .map_err(|e| cosmwasm_std::StdError::GenericErr { msg: e })?
                    .data
                    .ok_or_else(|| cosmwasm_std::StdError::NotFound {
                        kind: "cosmwasm_std::SubMsgResult::<T>".to_string(),
                    })?
                    .try_into()
            }
        }
    }).into()
}

fn get_type_url(attrs: &Vec<syn::Attribute>) -> proc_macro2::TokenStream {
    let proto_message = get_attr("proto_message", attrs).and_then(|a| a.parse_meta().ok());

    if let Some(syn::Meta::List(meta)) = proto_message.clone() {
        match meta.nested[0].clone() {
            syn::NestedMeta::Meta(syn::Meta::NameValue(meta)) => {
                if meta.path.is_ident("type_url") {
                    match meta.lit {
                        syn::Lit::Str(s) => quote!(#s),
                        _ => proto_message_attr_error(meta.lit),
                    }
                } else {
                    proto_message_attr_error(meta.path)
                }
            }
            t => proto_message_attr_error(t),
        }
    } else {
        proto_message_attr_error(proto_message)
    }
}

fn get_query_attrs<F>(attrs: &Vec<syn::Attribute>, f: F) -> proc_macro2::TokenStream
where
    F: FnMut(&Vec<TokenTree>) -> Option<proc_macro2::TokenStream>,
{
    let proto_query = get_attr("proto_query", attrs);

    if let Some(attr) = proto_query {
        if attr.tokens.clone().into_iter().count() != 1 {
            return proto_query_attr_error(proto_query);
        }

        if let Some(TokenTree::Group(group)) = attr.tokens.clone().into_iter().next() {
            let kv_groups = group.stream().into_iter().group_by(|t| {
                if let TokenTree::Punct(punct) = t {
                    punct.as_char() != ','
                } else {
                    true
                }
            });
            let mut key_values: Vec<Vec<TokenTree>> = vec![];

            for (non_sep, g) in &kv_groups {
                if non_sep {
                    key_values.push(g.collect());
                }
            }

            return key_values
                .iter()
                .find_map(f)
                .unwrap_or_else(|| proto_query_attr_error(proto_query));
        }

        proto_query_attr_error(proto_query)
    } else {
        proto_query_attr_error(proto_query)
    }
}

fn get_attr<'a>(attr_ident: &str, attrs: &'a Vec<syn::Attribute>) -> Option<&'a syn::Attribute> {
    attrs.iter().find(|&attr| attr.path.segments.len() == 1 && attr.path.segments[0].ident == attr_ident)
}

fn proto_message_attr_error<T: quote::ToTokens>(tokens: T) -> proc_macro2::TokenStream {
    syn::Error::new_spanned(tokens, "expected `proto_message(type_url = \"...\")`")
        .to_compile_error()
}

fn proto_query_attr_error<T: quote::ToTokens>(tokens: T) -> proc_macro2::TokenStream {
    syn::Error::new_spanned(
        tokens,
        "expected `proto_query(path = \"...\", response_type = ...)`",
    )
    .to_compile_error()
}
